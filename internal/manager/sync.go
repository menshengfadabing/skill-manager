package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncResult summarizes a sync run.
type SyncResult struct {
	Migrated []string
	Updated  []string
	Skipped  []string
	Warnings []string
	Enabled  []string
}

// Sync ingests real skill dirs / old links from both activity sets into the warehouse,
// then rebuilds dual direct symlinks for skills that were enabled on the agents side
// (union with previously warehouse-linked names on either side).
func (m *Manager) Sync() (*SyncResult, error) {
	if err := m.P.EnsureDirs(); err != nil {
		return nil, err
	}
	res := &SyncResult{}

	// Capture intended enabled set before mutation (agents authoritative, union claude links)
	agentsEnabled, err := m.EnabledOn(m.P.AgentsActive)
	if err != nil {
		return nil, err
	}
	claudeEnabled, err := m.EnabledOn(m.P.ClaudeActive)
	if err != nil {
		return nil, err
	}
	wantEnabled := map[string]struct{}{}
	for _, n := range agentsEnabled {
		wantEnabled[n] = struct{}{}
	}
	// Also keep names already correctly linked on either side
	for _, n := range claudeEnabled {
		wantEnabled[n] = struct{}{}
	}

	for _, active := range m.P.ActiveTargets() {
		if err := m.ingestActive(active, res); err != nil {
			return res, err
		}
	}

	// Rebuild enabled set as dual direct links
	var enableList []string
	for n := range wantEnabled {
		if isSkillDir(filepath.Join(m.P.Warehouse, n)) {
			enableList = append(enableList, n)
		} else {
			res.Warnings = append(res.Warnings, fmt.Sprintf("enabled name %q missing from warehouse after ingest", n))
		}
	}
	if err := m.ApplySet(enableList); err != nil {
		return res, err
	}
	res.Enabled = enableList

	// Seed default profile if missing
	def := filepath.Join(m.P.ProfilesDir, "default.txt")
	if _, err := os.Stat(def); os.IsNotExist(err) {
		if err := m.SaveProfile("default", enableList); err != nil {
			res.Warnings = append(res.Warnings, err.Error())
		}
	}
	if cur, _ := m.CurrentProfile(); cur == "" {
		_ = m.SetCurrentProfile("default")
	}
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

	// Two-level or other link: if target is a real skill dir, migrate it
	tinfo, err := os.Lstat(target)
	if err == nil && tinfo.Mode()&os.ModeSymlink == 0 && tinfo.IsDir() && isSkillDir(target) {
		// Only migrate if target is inside an activity set (not already warehouse)
		wh := filepath.Clean(m.P.Warehouse)
		if target != filepath.Join(wh, name) && !strings.HasPrefix(target, wh+string(os.PathSeparator)) {
			if err := m.migrateDir(target, name, res); err != nil {
				return err
			}
		}
	}
	// Replace link with direct warehouse link later via ApplySet; remove old link now
	if err := os.Remove(linkPath); err != nil {
		return err
	}
	res.Updated = append(res.Updated, linkPath)
	return nil
}

func (m *Manager) migrateDir(src, name string, res *SyncResult) error {
	dst := filepath.Join(m.P.Warehouse, name)
	if _, err := os.Lstat(dst); err == nil {
		// Conflict: keep newer tree
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
		backup := dst + ".bak-" + time.Now().Format("20060102-150405")
		res.Warnings = append(res.Warnings, fmt.Sprintf("conflict %s: replacing warehouse with newer %s (old → %s)", name, src, backup))
		if err := os.Rename(dst, backup); err != nil {
			return err
		}
	}
	if err := MoveDir(src, dst); err != nil {
		return err
	}
	res.Migrated = append(res.Migrated, name)
	return nil
}
