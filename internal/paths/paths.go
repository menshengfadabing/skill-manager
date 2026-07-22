package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkDir is one tool-facing skills directory we manage (启用集 / 工作目录).
type WorkDir struct {
	ID   string // agents | claude | cursor | codex | qwen | pi
	Path string
	Note string
}

// Paths is always user-global. Project dirs are ingest-only (迁移残留)，不作为启用写入目标。
type Paths struct {
	Scope      string // always "global"
	Root       string // home
	Home       string
	AgentsHome string
	Warehouse  string
	WorkDirs   []WorkDir

	AgentsActive string
	ClaudeActive string

	// ExtraIngest: optional project-level dirs to scan during sync only (do not ApplySet here).
	ExtraIngest []string

	ProfilesDir string
	CurrentFile string
	BackupsDir  string
}

// SupportedTools is the v1 tool set.
var SupportedTools = []string{
	"Claude Code",
	"Codex",
	"Cursor",
	"Qwen Code",
	"Pi",
}

// Resolve always returns the global layout. cwd is only used to discover project
// skill dirs for one-way sync ingest (avoid leaving orphans under a repo).
// The unused global flag is kept for CLI flag compatibility and ignored.
func Resolve(cwd string, _ bool) (Paths, error) {
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
	p := globalPaths(home)
	p.ExtraIngest = projectIngestDirs(cwd)
	return p, nil
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
	agentsActive := filepath.Join(agents, "skills")
	work := []WorkDir{
		{ID: "agents", Path: agentsActive, Note: "共享枢纽 · Codex / Cursor / Qwen / Pi"},
		{ID: "claude", Path: claude, Note: "Claude Code"},
		{ID: "cursor", Path: filepath.Join(home, ".cursor", "skills"), Note: "Cursor（不管 skills-cursor）"},
		{ID: "codex", Path: filepath.Join(home, ".codex", "skills"), Note: "Codex 遗留"},
		{ID: "qwen", Path: filepath.Join(home, ".qwen", "skills"), Note: "Qwen Code"},
		{ID: "pi", Path: filepath.Join(home, ".pi", "agent", "skills"), Note: "Pi"},
	}
	return Paths{
		Scope:        "global",
		Root:         home,
		Home:         home,
		AgentsHome:   agents,
		Warehouse:    sharedWarehouse(home),
		WorkDirs:     work,
		AgentsActive: agentsActive,
		ClaudeActive: claude,
		ProfilesDir:  profiles,
		CurrentFile:  filepath.Join(profiles, ".current"),
		BackupsDir:   filepath.Join(agents, "backups"),
	}
}

// projectIngestDirs lists repo-local skill dirs for sync ingest only.
func projectIngestDirs(cwd string) []string {
	root, err := findGitRoot(cwd)
	if err != nil {
		return nil
	}
	candidates := []string{
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".cursor", "skills"),
		filepath.Join(root, ".codex", "skills"),
		filepath.Join(root, ".qwen", "skills"),
		filepath.Join(root, ".pi", "skills"),
	}
	var out []string
	for _, d := range candidates {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			out = append(out, d)
		}
	}
	return out
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

// ActiveTargets: directories we write enable/disable into (global only).
func (p Paths) ActiveTargets() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	for _, w := range p.WorkDirs {
		add(w.Path)
	}
	if len(out) == 0 {
		add(p.AgentsActive)
		add(p.ClaudeActive)
	}
	return out
}

// IngestTargets: ActiveTargets plus optional project leftovers for sync.
func (p Paths) IngestTargets() []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	for _, d := range p.ActiveTargets() {
		add(d)
	}
	for _, d := range p.ExtraIngest {
		add(d)
	}
	return out
}

// RelLinkTo returns a relative symlink from activeDir to an absolute skill directory.
func (p Paths) RelLinkTo(activeDir, skillDir string) (string, error) {
	return filepath.Rel(activeDir, skillDir)
}

// EnsureDirs creates warehouse, global work dirs, profiles, and backups.
func (p Paths) EnsureDirs() error {
	dirs := []string{p.Warehouse, p.ProfilesDir, p.BackupsDir}
	for _, w := range p.WorkDirs {
		dirs = append(dirs, w.Path)
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Summary returns human-readable path lines.
func (p Paths) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "mode=global-only  home=%s\n", p.Home)
	fmt.Fprintf(&b, "warehouse=%s\n", p.Warehouse)
	fmt.Fprintf(&b, "tools=%s\n", strings.Join(SupportedTools, " / "))
	fmt.Fprintf(&b, "启用写入（全局工作目录）:\n")
	for _, w := range p.WorkDirs {
		fmt.Fprintf(&b, "  [%s] %s\n      %s\n", w.ID, w.Path, w.Note)
	}
	if len(p.ExtraIngest) > 0 {
		fmt.Fprintf(&b, "sync 额外摄入（仅迁移，不写回项目）:\n")
		for _, d := range p.ExtraIngest {
			fmt.Fprintf(&b, "  %s\n", d)
		}
	}
	fmt.Fprintf(&b, "profiles=%s\n", p.ProfilesDir)
	return strings.TrimRight(b.String(), "\n")
}
