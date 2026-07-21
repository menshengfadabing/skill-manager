package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths holds skill-manager filesystem layout for one scope (project or global).
//
// Warehouse is always ~/.agents/skills-all (shared). Project scope only owns
// activity sets + profiles under the repo; skill bodies live in the global warehouse.
type Paths struct {
	Scope        string // "project" or "global"
	Root         string // git root (project) or home (global)
	Home         string // user home
	AgentsHome   string // <repo>/.agents or ~/.agents (profiles/backups/activity for this scope)
	Warehouse    string // always ~/.agents/skills-all (shared true source)
	AgentsActive string // scope activity: .../skills
	ClaudeActive string // scope Claude mirror
	ProfilesDir  string
	CurrentFile  string
	BackupsDir   string
}

// Resolve returns paths for project (default) or global (-g) scope.
func Resolve(cwd string, global bool) (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return Paths{}, err
		}
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return Paths{}, err
	}

	if global {
		return globalPaths(home), nil
	}

	root, err := findGitRoot(cwd)
	if err != nil {
		return Paths{}, fmt.Errorf("未找到 git 仓库（当前目录: %s）。项目级操作需在仓库内执行；全局请加 -g/--global", cwd)
	}
	return projectPaths(home, root), nil
}

func sharedWarehouse(home string) string {
	agents := filepath.Join(home, ".agents")
	if v := os.Getenv("SKILL_MANAGER_HOME"); v != "" {
		agents = v
	}
	return filepath.Join(agents, "skills-all")
}

func globalPaths(home string) Paths {
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
		Scope:        "global",
		Root:         home,
		Home:         home,
		AgentsHome:   agents,
		Warehouse:    sharedWarehouse(home),
		AgentsActive: filepath.Join(agents, "skills"),
		ClaudeActive: claude,
		ProfilesDir:  profiles,
		CurrentFile:  filepath.Join(profiles, ".current"),
		BackupsDir:   filepath.Join(agents, "backups"),
	}
}

func projectPaths(home, root string) Paths {
	agents := filepath.Join(root, ".agents")
	profiles := filepath.Join(agents, "profiles")
	return Paths{
		Scope:        "project",
		Root:         root,
		Home:         home,
		AgentsHome:   agents,
		Warehouse:    sharedWarehouse(home), // shared with global — no per-project skills-all
		AgentsActive: filepath.Join(agents, "skills"),
		ClaudeActive: filepath.Join(root, ".claude", "skills"),
		ProfilesDir:  profiles,
		CurrentFile:  filepath.Join(profiles, ".current"),
		BackupsDir:   filepath.Join(agents, "backups"),
	}
}

func findGitRoot(start string) (string, error) {
	dir := start
	for {
		if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil && (st.IsDir() || st.Mode().IsRegular()) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git root")
		}
		dir = parent
	}
}

// ActiveTargets: agents (Cursor/Codex/Qwen) + claude (CC only).
func (p Paths) ActiveTargets() []string {
	return []string{p.AgentsActive, p.ClaudeActive}
}

// RelLinkTo returns a relative symlink from activeDir to an absolute skill directory.
func (p Paths) RelLinkTo(activeDir, skillDir string) (string, error) {
	return filepath.Rel(activeDir, skillDir)
}

// EnsureDirs creates warehouse, actives, profiles, and backups directories.
func (p Paths) EnsureDirs() error {
	for _, d := range []string{p.Warehouse, p.AgentsActive, p.ClaudeActive, p.ProfilesDir, p.BackupsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Summary returns human-readable path lines for list/doctor headers.
func (p Paths) Summary() string {
	return fmt.Sprintf("scope=%s root=%s\nwarehouse=%s  (shared global)\nagents=%s  (Cursor/Codex/Qwen)\nclaude=%s  (Claude Code only)\nprofiles=%s",
		p.Scope, p.Root, p.Warehouse, p.AgentsActive, p.ClaudeActive, p.ProfilesDir)
}
