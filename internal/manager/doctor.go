package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Issue is one doctor finding.
type Issue struct {
	Level   string // error, warn, info
	Path    string
	Message string
}

// Doctor scans for two-level links, real dirs in actives, one-sided enable, broken links.
func (m *Manager) Doctor() ([]Issue, error) {
	var issues []Issue
	if err := m.P.EnsureDirs(); err != nil {
		return nil, err
	}

	targets := m.P.ActiveTargets()
	sets := make([]map[string]struct{}, len(targets))
	labels := []string{"agents", "claude"}
	for i, active := range targets {
		names, err := m.EnabledOn(active)
		if err != nil {
			return nil, err
		}
		sets[i] = toSet(names)
	}
	// pairwise drift vs agents (index 0)
	for i := 1; i < len(sets); i++ {
		for name := range sets[0] {
			if _, ok := sets[i][name]; !ok {
				issues = append(issues, Issue{Level: "warn", Path: name, Message: fmt.Sprintf("enabled on agents but missing on %s", labels[i])})
			}
		}
		for name := range sets[i] {
			if _, ok := sets[0][name]; !ok {
				issues = append(issues, Issue{Level: "warn", Path: name, Message: fmt.Sprintf("enabled on %s but missing on agents", labels[i])})
			}
		}
	}

	for _, active := range targets {
		entries, err := os.ReadDir(active)
		if err != nil {
			continue
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
				if m.IsTwoLevelLink(full) {
					issues = append(issues, Issue{Level: "error", Path: full, Message: "two-level symlink (points into .agents/skills); should link skills-all directly"})
				} else if !m.IsDirectWarehouseLink(full, name) {
					target, _ := os.Readlink(full)
					issues = append(issues, Issue{Level: "warn", Path: full, Message: fmt.Sprintf("symlink not direct to warehouse (→ %s)", target)})
				}
				if _, err := os.Stat(full); err != nil {
					issues = append(issues, Issue{Level: "error", Path: full, Message: "broken symlink"})
				}
			} else if info.IsDir() && isSkillDir(full) {
				issues = append(issues, Issue{Level: "error", Path: full, Message: "real skill directory in activity set; run skill sync"})
			}
		}
	}

	whEntries, _ := os.ReadDir(m.P.Warehouse)
	for _, e := range whEntries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(m.P.Warehouse, e.Name())
		if !isSkillDir(p) {
			issues = append(issues, Issue{Level: "warn", Path: p, Message: "warehouse dir missing SKILL.md"})
		}
	}

	if len(issues) == 0 {
		issues = append(issues, Issue{Level: "info", Message: "no issues found"})
	}
	return issues, nil
}

func toSet(names []string) map[string]struct{} {
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return m
}
