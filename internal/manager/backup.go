package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/no-story/skill-manager/internal/paths"
)

const (
	// InitialSnapshotID is the baseline taken before the first destructive sync/init.
	InitialSnapshotID = "INITIAL"
)

// Snapshot describes one backup under BackupsDir.
type Snapshot struct {
	ID        string
	Path      string
	Label     string
	IsInitial bool
	Created   time.Time
}

// BackupSnapshot copies mutable scope state into AgentsHome/backups/<ts>[-label]/.
func (m *Manager) BackupSnapshot(label string) (string, error) {
	if err := os.MkdirAll(m.P.BackupsDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	if label != "" {
		ts = ts + "-" + label
	}
	dst := filepath.Join(m.P.BackupsDir, ts)
	if err := m.writeSnapshot(dst, label, false); err != nil {
		return dst, err
	}
	return dst, nil
}

// EnsureInitialSnapshot writes INITIAL once, before any destructive change.
func (m *Manager) EnsureInitialSnapshot() (string, error) {
	dst := filepath.Join(m.P.BackupsDir, InitialSnapshotID)
	if st, err := os.Stat(dst); err == nil && st.IsDir() {
		return dst, nil
	}
	if err := os.MkdirAll(m.P.BackupsDir, 0o755); err != nil {
		return "", err
	}
	if err := m.writeSnapshot(dst, "initial", true); err != nil {
		return dst, err
	}
	return dst, nil
}

func (m *Manager) writeSnapshot(dst, label string, initial bool) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	kind := "auto"
	if initial {
		kind = "initial"
	}
	manifest := fmt.Sprintf("scope=%s\nroot=%s\nwarehouse=%s\nlabel=%s\nkind=%s\n",
		m.P.Scope, m.P.Root, m.P.Warehouse, label, kind)
	if err := os.WriteFile(filepath.Join(dst, "MANIFEST.txt"), []byte(manifest+m.P.Summary()+"\n"), 0o644); err != nil {
		return err
	}
	_ = os.WriteFile(filepath.Join(dst, "LABEL.txt"), []byte(label+"\n"), 0o644)
	if initial {
		_ = os.WriteFile(filepath.Join(dst, "INITIAL.marker"), []byte("user baseline before skill sync/init\n"), 0o644)
	}

	if err := copyIfExists(m.P.ProfilesDir, filepath.Join(dst, "profiles")); err != nil {
		return err
	}
	work := m.P.WorkDirs
	if len(work) == 0 {
		work = []paths.WorkDir{
			{ID: "agents", Path: m.P.AgentsActive},
			{ID: "claude", Path: m.P.ClaudeActive},
		}
	}
	for _, w := range work {
		if w.Path == "" {
			continue
		}
		list, _ := m.EnabledOn(w.Path)
		var b string
		for _, n := range list {
			b += n + "\n"
		}
		_ = os.WriteFile(filepath.Join(dst, w.ID+"-enabled.txt"), []byte(b), 0o644)
		if err := copyIfExists(w.Path, filepath.Join(dst, w.ID+"-skills")); err != nil {
			return err
		}
	}

	names, _ := m.SkillNames()
	var wb string
	for _, n := range names {
		wb += n + "\n"
	}
	_ = os.WriteFile(filepath.Join(dst, "warehouse-names.txt"), []byte(wb), 0o644)
	return nil
}

// ListSnapshots returns backups newest-first; INITIAL always listed (if present) with IsInitial.
func (m *Manager) ListSnapshots() ([]Snapshot, error) {
	entries, err := os.ReadDir(m.P.BackupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Snapshot
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		path := filepath.Join(m.P.BackupsDir, id)
		info, err := e.Info()
		if err != nil {
			continue
		}
		label := ""
		if b, err := os.ReadFile(filepath.Join(path, "LABEL.txt")); err == nil {
			label = strings.TrimSpace(string(b))
		}
		isInit := id == InitialSnapshotID
		if _, err := os.Stat(filepath.Join(path, "INITIAL.marker")); err == nil {
			isInit = true
		}
		created := info.ModTime()
		out = append(out, Snapshot{
			ID:        id,
			Path:      path,
			Label:     label,
			IsInitial: isInit,
			Created:   created,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsInitial != out[j].IsInitial {
			return out[i].IsInitial // initial first
		}
		return out[i].Created.After(out[j].Created)
	})
	return out, nil
}

// ResolveSnapshotID accepts "initial"/"INITIAL"/exact folder name/prefix.
func (m *Manager) ResolveSnapshotID(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("empty snapshot id")
	}
	low := strings.ToLower(ref)
	if low == "initial" || low == "init-baseline" || ref == InitialSnapshotID {
		path := filepath.Join(m.P.BackupsDir, InitialSnapshotID)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("还没有「用户初始」快照（首次 sync/init 前才会自动创建）")
		}
		return InitialSnapshotID, nil
	}
	exact := filepath.Join(m.P.BackupsDir, ref)
	if st, err := os.Stat(exact); err == nil && st.IsDir() {
		return ref, nil
	}
	entries, err := os.ReadDir(m.P.BackupsDir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), ref) {
			matches = append(matches, e.Name())
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("快照 id 不唯一，匹配到: %s", strings.Join(matches, ", "))
	}
	return "", fmt.Errorf("找不到快照 %q（skill log 查看列表）", ref)
}

// RestoreSnapshot replaces profiles + working dirs from a snapshot.
func (m *Manager) RestoreSnapshot(ref string) error {
	id, err := m.ResolveSnapshotID(ref)
	if err != nil {
		return err
	}
	src := filepath.Join(m.P.BackupsDir, id)
	if err := m.P.EnsureDirs(); err != nil {
		return err
	}

	// Safety backup of current state before restore
	if _, err := m.BackupSnapshot("pre-restore"); err != nil {
		return fmt.Errorf("restore 前备份失败: %w", err)
	}

	if err := replaceDir(filepath.Join(src, "profiles"), m.P.ProfilesDir); err != nil {
		return fmt.Errorf("restore profiles: %w", err)
	}
	work := m.P.WorkDirs
	if len(work) == 0 {
		work = []paths.WorkDir{
			{ID: "agents", Path: m.P.AgentsActive},
			{ID: "claude", Path: m.P.ClaudeActive},
		}
	}
	for _, w := range work {
		if w.Path == "" {
			continue
		}
		bak := filepath.Join(src, w.ID+"-skills")
		if err := replaceDir(bak, w.Path); err != nil {
			return fmt.Errorf("restore %s working dir: %w", w.ID, err)
		}
	}
	return nil
}

func replaceDir(backupSrc, liveDst string) error {
	if err := os.RemoveAll(liveDst); err != nil {
		return err
	}
	if _, err := os.Stat(backupSrc); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(liveDst, 0o755)
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(liveDst), 0o755); err != nil {
		return err
	}
	return CopyDir(backupSrc, liveDst)
}

// UninstallOptions controls skill uninstall.
type UninstallOptions struct {
	RestoreInitial bool
	Yes            bool
}

// Uninstall restores optional INITIAL, then removes skill-manager footprint in this scope.
func (m *Manager) Uninstall(opts UninstallOptions) error {
	if opts.RestoreInitial {
		if err := m.RestoreSnapshot(InitialSnapshotID); err != nil {
			return fmt.Errorf("恢复用户初始失败: %w", err)
		}
	}

	// Remove companion skills from working dirs only (not entire warehouse of other skills)
	for _, active := range m.P.ActiveTargets() {
		for _, name := range []string{"skill-manager", "skill-init"} {
			_ = os.RemoveAll(filepath.Join(active, name))
		}
	}
	// Remove companion skill bodies we may have installed into warehouse
	for _, name := range []string{"skill-manager", "skill-init"} {
		_ = os.RemoveAll(filepath.Join(m.P.Warehouse, name))
	}

	// Clean skill-manager-managed dirs in this scope
	for _, d := range []string{m.P.ProfilesDir, m.P.BackupsDir} {
		_ = os.RemoveAll(d)
	}
	// Remove empty .agents/skills / .claude/skills only if empty after companion removal? Keep dirs.
	// Drop AgentsHome leftovers that we created: if AgentsHome only had profiles/backups/skills, leave skills.
	return nil
}

func copyIfExists(src, dst string) error {
	st, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 || !st.IsDir() {
		return CopyDir(src, dst)
	}
	return CopyDir(src, dst)
}
