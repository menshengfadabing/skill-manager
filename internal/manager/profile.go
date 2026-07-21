package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile is the on-disk YAML profile document.
type Profile struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Skills      []string `yaml:"skills"`
}

func (m *Manager) profilePath(name string) string {
	return filepath.Join(m.P.ProfilesDir, name+".yaml")
}

// ReadProfile loads a profile skill list from YAML.
func (m *Manager) ReadProfile(name string) ([]string, error) {
	p, err := m.LoadProfile(name)
	if err != nil {
		return nil, err
	}
	return p.Skills, nil
}

// LoadProfile returns the full profile document.
func (m *Manager) LoadProfile(name string) (*Profile, error) {
	b, err := os.ReadFile(m.profilePath(name))
	if err != nil {
		return nil, fmt.Errorf("profile %q: %w", name, err)
	}
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("profile %q: %w", name, err)
	}
	if p.Name == "" {
		p.Name = name
	}
	return &p, nil
}

// WriteProfile writes profiles/<name>.yaml.
func (m *Manager) WriteProfile(p *Profile) error {
	if err := os.MkdirAll(m.P.ProfilesDir, 0o755); err != nil {
		return err
	}
	if p.Name == "" {
		return fmt.Errorf("profile name required")
	}
	if p.Skills == nil {
		p.Skills = []string{}
	}
	b, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(m.profilePath(p.Name), b, 0o644)
}

// SaveProfile writes skills into an existing-or-new named profile (TUI/internal).
func (m *Manager) SaveProfile(name string, skills []string) error {
	desc := ""
	if existing, err := m.LoadProfile(name); err == nil {
		desc = existing.Description
	}
	return m.WriteProfile(&Profile{Name: name, Description: desc, Skills: skills})
}

// CreateProfile creates an empty YAML profile.
func (m *Manager) CreateProfile(name, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("profile name required")
	}
	if _, err := os.Stat(m.profilePath(name)); err == nil {
		return fmt.Errorf("profile %q already exists", name)
	}
	return m.WriteProfile(&Profile{Name: name, Description: description, Skills: []string{}})
}

// DeleteProfile removes a profile file.
func (m *Manager) DeleteProfile(name string, force bool) error {
	cur, _ := m.CurrentProfile()
	if cur == name && !force {
		return fmt.Errorf("拒绝删除当前正在使用的 profile %q（加 --force 强制）", name)
	}
	yp := m.profilePath(name)
	if err := os.Remove(yp); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile %q not found", name)
		}
		return err
	}
	if cur == name {
		_ = os.Remove(m.P.CurrentFile)
	}
	return nil
}

// UseProfile applies a named profile to all activity targets in this scope.
func (m *Manager) UseProfile(name string) error {
	skills, err := m.ReadProfile(name)
	if err != nil {
		return err
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

// CoreSkills is the minimal activity set for skill init.
var CoreSkills = []string{"skill-manager", "skill-init"}

// EnsureCoreProfile writes/refreshes the core profile to the canonical minimal set.
func (m *Manager) EnsureCoreProfile() error {
	var filtered []string
	for _, n := range CoreSkills {
		if _, err := m.ResolveSkillDir(n); err == nil {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		filtered = append([]string{}, CoreSkills...)
	}
	return m.WriteProfile(&Profile{
		Name:        "core",
		Description: "最小集：skill-manager + skill-init",
		Skills:      filtered,
	})
}

// Init refreshes core profile and switches to it.
func (m *Manager) Init() error {
	if err := m.EnsureCoreProfile(); err != nil {
		return err
	}
	return m.UseProfile("core")
}

// ListProfiles returns profile names (.yaml only).
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
		if strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return names, nil
}
