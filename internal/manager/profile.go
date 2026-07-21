package manager

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadProfile loads a profile name list (one skill per line, # comments ok).
func (m *Manager) ReadProfile(name string) ([]string, error) {
	path := filepath.Join(m.P.ProfilesDir, name+".txt")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var skills []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		skills = append(skills, line)
	}
	return skills, sc.Err()
}

// SaveProfile writes the enabled list to profiles/<name>.txt.
func (m *Manager) SaveProfile(name string, skills []string) error {
	if err := os.MkdirAll(m.P.ProfilesDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(m.P.ProfilesDir, name+".txt")
	var b strings.Builder
	b.WriteString("# skill-manager profile: " + name + "\n")
	for _, s := range skills {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		b.WriteString(s)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// UseProfile applies a named profile to both activity sets.
func (m *Manager) UseProfile(name string) error {
	skills, err := m.ReadProfile(name)
	if err != nil {
		return fmt.Errorf("profile %q: %w", name, err)
	}
	if err := m.ApplySet(skills); err != nil {
		return err
	}
	return m.SetCurrentProfile(name)
}

// CurrentProfile returns the active profile name.
func (m *Manager) CurrentProfile() (string, error) {
	b, err := os.ReadFile(m.P.CurrentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// SetCurrentProfile records the current profile name.
func (m *Manager) SetCurrentProfile(name string) error {
	if err := os.MkdirAll(m.P.ProfilesDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(m.P.CurrentFile, []byte(name+"\n"), 0o644)
}

// EnsureCoreProfile writes the core profile if missing.
func (m *Manager) EnsureCoreProfile() error {
	path := filepath.Join(m.P.ProfilesDir, "core.txt")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	core := []string{"skill-manager", "skill-init"}
	// Only include names that exist in warehouse
	var filtered []string
	for _, n := range core {
		if isSkillDir(filepath.Join(m.P.Warehouse, n)) {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		filtered = core // still write; use will error until bundled install
	}
	return m.SaveProfile("core", filtered)
}

// Init switches to core profile (minimal activity set).
func (m *Manager) Init() error {
	if err := m.EnsureCoreProfile(); err != nil {
		return err
	}
	return m.UseProfile("core")
}

// ListProfiles returns profile names (without .txt).
func (m *Manager) ListProfiles() ([]string, error) {
	entries, err := os.ReadDir(m.P.ProfilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.HasSuffix(e.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(e.Name(), ".txt"))
		}
	}
	return names, nil
}
