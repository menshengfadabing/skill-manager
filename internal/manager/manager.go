package manager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/no-story/skill-manager/internal/paths"
)

// Manager operates on the skills warehouse and dual activity sets.
type Manager struct {
	P paths.Paths
}

func New(p paths.Paths) *Manager {
	return &Manager{P: p}
}

// SkillNames lists skills in the shared global warehouse.
func (m *Manager) SkillNames() ([]string, error) {
	entries, err := os.ReadDir(m.P.Warehouse)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !isSkillDir(filepath.Join(m.P.Warehouse, name)) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// ResolveSkillDir returns the absolute skill directory in the shared warehouse.
func (m *Manager) ResolveSkillDir(name string) (string, error) {
	dir := filepath.Join(m.P.Warehouse, name)
	if isSkillDir(dir) {
		return dir, nil
	}
	return "", fmt.Errorf("skill %q not in warehouse %s", name, m.P.Warehouse)
}

func isSkillDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "SKILL.md"))
	return err == nil
}

// EnabledOn returns skill names currently present in an activity directory.
func (m *Manager) EnabledOn(activeDir string) ([]string, error) {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(activeDir, name)
		// Include dirs and symlinks that resolve to skill-like paths
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// EnabledAgents returns names enabled under .agents/skills.
func (m *Manager) EnabledAgents() ([]string, error) {
	return m.EnabledOn(m.P.AgentsActive)
}

// ResolveWarehouseTarget returns the absolute target if link points into the shared warehouse.
func (m *Manager) ResolveWarehouseTarget(linkPath string) (string, bool) {
	info, err := os.Lstat(linkPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return "", false
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		return "", false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	target = filepath.Clean(target)
	wh := filepath.Clean(m.P.Warehouse)
	if target == wh || strings.HasPrefix(target, wh+string(os.PathSeparator)) {
		return target, true
	}
	return target, false
}

// IsDirectWarehouseLink reports whether path is a symlink directly to the resolved skill dir.
func (m *Manager) IsDirectWarehouseLink(linkPath, name string) bool {
	target, ok := m.ResolveWarehouseTarget(linkPath)
	if !ok {
		return false
	}
	want, err := m.ResolveSkillDir(name)
	if err != nil {
		return false
	}
	return filepath.Clean(target) == filepath.Clean(want)
}

// IsTwoLevelLink detects .claude → .agents/skills → ... style chains.
func (m *Manager) IsTwoLevelLink(linkPath string) bool {
	info, err := os.Lstat(linkPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	target = filepath.Clean(target)
	agents := filepath.Clean(m.P.AgentsActive)
	// Points into agents active (not warehouse)
	if target == agents || strings.HasPrefix(target, agents+string(os.PathSeparator)) {
		return true
	}
	return false
}

// Enable creates direct warehouse symlinks on all activity sets.
func (m *Manager) Enable(names ...string) error {
	if err := m.P.EnsureDirs(); err != nil {
		return err
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, err := m.ResolveSkillDir(name); err != nil {
			return err
		}
		for _, active := range m.P.ActiveTargets() {
			if err := m.linkOne(active, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) linkOne(activeDir, name string) error {
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		return err
	}
	skillDir, err := m.ResolveSkillDir(name)
	if err != nil {
		return err
	}
	linkPath := filepath.Join(activeDir, name)
	rel, err := m.P.RelLinkTo(activeDir, skillDir)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(linkPath); err == nil {
		info, _ := os.Lstat(linkPath)
		if info.Mode()&os.ModeSymlink == 0 && info.IsDir() {
			return fmt.Errorf("%s is a real directory; run 'skill sync' first", linkPath)
		}
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	}
	return os.Symlink(rel, linkPath)
}

// Disable removes skill symlinks from both activity sets (does not touch warehouse).
func (m *Manager) Disable(names ...string) error {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		for _, active := range m.P.ActiveTargets() {
			linkPath := filepath.Join(active, name)
			info, err := os.Lstat(linkPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("%s is not a symlink; run 'skill sync' first", linkPath)
			}
			if err := os.Remove(linkPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// ApplySet makes both activity sets match exactly the given enabled names
// (only manages warehouse-backed symlinks; refuses if real dirs remain).
func (m *Manager) ApplySet(names []string) error {
	if err := m.P.EnsureDirs(); err != nil {
		return err
	}
	want := map[string]struct{}{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		want[n] = struct{}{}
		if _, err := m.ResolveSkillDir(n); err != nil {
			return fmt.Errorf("skill %q not in warehouse %s; run skill sync (-g if needed)", n, m.P.Warehouse)
		}
	}
	for _, active := range m.P.ActiveTargets() {
		if err := m.ensureNoRealSkills(active); err != nil {
			return err
		}
		current, err := m.EnabledOn(active)
		if err != nil {
			return err
		}
		for _, name := range current {
			linkPath := filepath.Join(active, name)
			info, err := os.Lstat(linkPath)
			if err != nil {
				continue
			}
			if info.Mode()&os.ModeSymlink == 0 {
				continue
			}
			if _, ok := want[name]; !ok {
				if err := os.Remove(linkPath); err != nil {
					return err
				}
			}
		}
		for name := range want {
			if err := m.linkOne(active, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) ensureNoRealSkills(activeDir string) error {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		full := filepath.Join(activeDir, e.Name())
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 && info.IsDir() && isSkillDir(full) {
			return fmt.Errorf("real skill dir at %s; run 'skill sync' before changing profiles", full)
		}
	}
	return nil
}

// CopyDir copies a directory tree (used when rename across devices fails).
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// MoveDir moves src to dst, falling back to copy+remove.
func MoveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := CopyDir(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}
	return os.RemoveAll(src)
}
