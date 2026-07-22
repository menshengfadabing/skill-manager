package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncOptions controls destructive sync behavior.
type SyncOptions struct {
	DryRun bool
	Yes    bool // skip confirm (caller should already have confirmed when false means need confirm)
}

// SyncResult summarizes a sync run.
type SyncResult struct {
	Migrated   []string
	Updated    []string
	Skipped    []string
	Warnings   []string
	Enabled    []string
	Bundled    []string
	BackupDir  string
	DryRun     bool
	WouldMigrate []string
	WouldUpdate  []string
}

// SyncPlan describes what sync would do without writing (for dry-run).
type SyncPlan struct {
	Migrate []string
	Update  []string
	RealDirs []string
}

// PlanSync inspects activity sets without mutating.
func (m *Manager) PlanSync() (*SyncPlan, error) {
	plan := &SyncPlan{}
	for _, active := range m.P.IngestTargets() {
		entries, err := os.ReadDir(active)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			full := filepath.Join(active, name)
			info, err := os.Lstat(full)
			if err != nil {
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				if !m.IsDirectWarehouseLink(full, name) {
					plan.Update = append(plan.Update, full)
				}
				continue
			}
			if info.IsDir() && isSkillDir(full) {
				plan.Migrate = append(plan.Migrate, name+" @"+full)
				plan.RealDirs = append(plan.RealDirs, full)
			}
		}
	}
	return plan, nil
}

// Sync ingests real skill dirs / old links into the warehouse, then rebuilds mirrors.
func (m *Manager) Sync(opts SyncOptions) (*SyncResult, error) {
	if err := m.P.EnsureDirs(); err != nil {
		return nil, err
	}
	res := &SyncResult{DryRun: opts.DryRun}

	if opts.DryRun {
		plan, err := m.PlanSync()
		if err != nil {
			return res, err
		}
		res.WouldMigrate = plan.Migrate
		res.WouldUpdate = plan.Update
		return res, nil
	}

	if _, err := m.EnsureInitialSnapshot(); err != nil {
		return res, fmt.Errorf("initial snapshot: %w", err)
	}
	bak, err := m.BackupSnapshot("sync")
	if err != nil {
		return res, fmt.Errorf("backup failed: %w", err)
	}
	res.BackupDir = bak

	bundled, err := m.InstallBundled()
	if err != nil {
		res.Warnings = append(res.Warnings, "bundled: "+err.Error())
	} else {
		res.Bundled = bundled
	}

	// Only global work dirs decide the enable set. ExtraIngest (project leftovers)
	// may contribute bodies to the warehouse, but must not expand global enables.
	wantEnabled := map[string]struct{}{}
	for _, active := range m.P.ActiveTargets() {
		enabled, err := m.EnabledOn(active)
		if err != nil {
			return res, err
		}
		for _, n := range enabled {
			wantEnabled[n] = struct{}{}
		}
	}

	for _, active := range m.P.IngestTargets() {
		if err := m.ingestActive(active, res); err != nil {
			return res, err
		}
	}

	var enableList []string
	for n := range wantEnabled {
		if _, err := m.ResolveSkillDir(n); err == nil {
			enableList = append(enableList, n)
		} else {
			res.Warnings = append(res.Warnings, fmt.Sprintf("enabled name %q missing from warehouse after ingest", n))
		}
	}
	if err := m.ApplySet(enableList); err != nil {
		return res, err
	}
	res.Enabled = enableList

	def := m.profilePath("default")
	if _, err := os.Stat(def); os.IsNotExist(err) {
		if err := m.SaveProfile("default", enableList); err != nil {
			res.Warnings = append(res.Warnings, err.Error())
		}
	}
	if cur, _ := m.CurrentProfile(); cur == "" {
		_ = m.SetCurrentProfile("default")
	}
	_ = m.EnsureCoreProfile()
	return res, nil
}

func (m *Manager) ingestActive(activeDir string, res *SyncResult) error {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(activeDir, name)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := m.ingestSymlink(full, name, res); err != nil {
				res.Warnings = append(res.Warnings, err.Error())
			}
			continue
		}
		if info.IsDir() {
			if !isSkillDir(full) {
				res.Skipped = append(res.Skipped, full)
				continue
			}
			if err := m.migrateDir(full, name, res); err != nil {
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s: %v", full, err))
			}
		}
	}
	return nil
}

func (m *Manager) ingestSymlink(linkPath, name string, res *SyncResult) error {
	if m.IsDirectWarehouseLink(linkPath, name) {
		res.Skipped = append(res.Skipped, linkPath+" (already direct)")
		return nil
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	target = filepath.Clean(target)

	tinfo, err := os.Lstat(target)
	if err == nil && tinfo.Mode()&os.ModeSymlink == 0 && tinfo.IsDir() && isSkillDir(target) {
		wh := filepath.Clean(m.P.Warehouse)
		if target != filepath.Join(wh, name) && !strings.HasPrefix(target, wh+string(os.PathSeparator)) {
			if err := m.migrateDir(target, name, res); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(linkPath); err != nil {
		return err
	}
	res.Updated = append(res.Updated, linkPath)
	return nil
}

func (m *Manager) migrateDir(src, name string, res *SyncResult) error {
	dst := filepath.Join(m.P.Warehouse, name)
	if _, err := os.Lstat(dst); err == nil {
		srcInfo, _ := os.Stat(src)
		dstInfo, _ := os.Stat(dst)
		if srcInfo != nil && dstInfo != nil && !srcInfo.ModTime().After(dstInfo.ModTime()) {
			res.Warnings = append(res.Warnings, fmt.Sprintf("conflict %s: warehouse newer, removing duplicate at %s", name, src))
			if err := os.RemoveAll(src); err != nil {
				return err
			}
			res.Skipped = append(res.Skipped, src)
			return nil
		}
		bakRoot := filepath.Join(filepath.Dir(m.P.Warehouse), "skills-bak")
		if err := os.MkdirAll(bakRoot, 0o755); err != nil {
			return err
		}
		backup := filepath.Join(bakRoot, name+"-"+time.Now().Format("20060102-150405"))
		res.Warnings = append(res.Warnings, fmt.Sprintf("conflict %s: replacing warehouse with newer %s (old → %s)", name, src, backup))
		if err := MoveDir(dst, backup); err != nil {
			return err
		}
	}
	if err := MoveDir(src, dst); err != nil {
		return err
	}
	res.Migrated = append(res.Migrated, name)
	return nil
}

// InstallBundled copies companion skills from known locations into warehouse if missing.
func (m *Manager) InstallBundled() ([]string, error) {
	srcRoot := findBundledRoot()
	if srcRoot == "" {
		return nil, fmt.Errorf("bundled skills directory not found")
	}
	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		return nil, err
	}
	var installed []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		src := filepath.Join(srcRoot, e.Name())
		dst := filepath.Join(m.P.Warehouse, e.Name())
		if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
			continue
		}
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := CopyDir(src, dst); err != nil {
			return installed, err
		}
		installed = append(installed, e.Name())
	}
	return installed, nil
}

func findBundledRoot() string {
	var candidates []string
	if v := os.Getenv("SKILL_MANAGER_BUNDLED"); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates, "/usr/local/share/skill-manager/skills")
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "skills"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "skills"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return ""
}
