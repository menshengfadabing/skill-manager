package paths

import (
	"os"
	"path/filepath"
)

// Paths holds skill-manager filesystem layout.
type Paths struct {
	Home         string // usually $HOME
	AgentsHome   string // ~/.agents
	Warehouse    string // ~/.agents/skills-all
	AgentsActive string // ~/.agents/skills
	ClaudeActive string // ~/.claude/skills
	ProfilesDir  string // ~/.agents/profiles
	CurrentFile  string // ~/.agents/profiles/.current
}

// Default returns paths under the current user home.
// Override with SKILL_MANAGER_HOME (agents root) and optionally
// SKILL_MANAGER_CLAUDE (claude skills dir).
func Default() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	agents := filepath.Join(home, ".agents")
	if v := os.Getenv("SKILL_MANAGER_HOME"); v != "" {
		agents = v
	}
	claude := filepath.Join(home, ".claude", "skills")
	if v := os.Getenv("SKILL_MANAGER_CLAUDE"); v != "" {
		claude = v
	}
	profiles := filepath.Join(agents, "profiles")
	return Paths{
		Home:         home,
		AgentsHome:   agents,
		Warehouse:    filepath.Join(agents, "skills-all"),
		AgentsActive: filepath.Join(agents, "skills"),
		ClaudeActive: claude,
		ProfilesDir:  profiles,
		CurrentFile:  filepath.Join(profiles, ".current"),
	}, nil
}

// ActiveTargets returns activity-set directories that receive symlinks.
func (p Paths) ActiveTargets() []string {
	return []string{p.AgentsActive, p.ClaudeActive}
}

// RelLink returns a relative symlink target from activeDir to warehouse/name.
func (p Paths) RelLink(activeDir, name string) (string, error) {
	dest := filepath.Join(p.Warehouse, name)
	return filepath.Rel(activeDir, dest)
}

// EnsureDirs creates warehouse, actives, and profiles directories.
func (p Paths) EnsureDirs() error {
	for _, d := range []string{p.Warehouse, p.AgentsActive, p.ClaudeActive, p.ProfilesDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
