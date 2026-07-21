package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/no-story/skill-manager/internal/manager"
	"github.com/no-story/skill-manager/internal/paths"
	"github.com/no-story/skill-manager/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	p, err := paths.Default()
	if err != nil {
		return err
	}
	m := manager.New(p)

	if len(args) == 0 {
		return cmdTUI(m)
	}
	switch args[0] {
	case "list", "ls":
		return cmdList(m)
	case "enable":
		if len(args) < 2 {
			return fmt.Errorf("usage: skill enable <name...>")
		}
		return m.Enable(args[1:]...)
	case "disable":
		if len(args) < 2 {
			return fmt.Errorf("usage: skill disable <name...>")
		}
		return m.Disable(args[1:]...)
	case "sync":
		res, err := m.Sync()
		if err != nil {
			return err
		}
		printSync(res)
		return nil
	case "doctor":
		issues, err := m.Doctor()
		if err != nil {
			return err
		}
		for _, iss := range issues {
			loc := iss.Path
			if loc != "" {
				loc = " " + loc
			}
			fmt.Printf("[%s]%s %s\n", iss.Level, loc, iss.Message)
		}
		return nil
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: skill use <profile>")
		}
		return m.UseProfile(args[1])
	case "save":
		if len(args) < 2 {
			return fmt.Errorf("usage: skill save <profile>")
		}
		enabled, err := m.EnabledAgents()
		if err != nil {
			return err
		}
		if err := m.SaveProfile(args[1], enabled); err != nil {
			return err
		}
		fmt.Printf("saved profile %q (%d skills)\n", args[1], len(enabled))
		return nil
	case "init":
		if err := m.Init(); err != nil {
			return err
		}
		fmt.Println("switched to core profile")
		return nil
	case "profiles":
		names, err := m.ListProfiles()
		if err != nil {
			return err
		}
		cur, _ := m.CurrentProfile()
		for _, n := range names {
			mark := " "
			if n == cur {
				mark = "*"
			}
			fmt.Printf("%s %s\n", mark, n)
		}
		return nil
	case "path", "paths":
		fmt.Printf("warehouse:     %s\n", p.Warehouse)
		fmt.Printf("agents active: %s\n", p.AgentsActive)
		fmt.Printf("claude active: %s\n", p.ClaudeActive)
		fmt.Printf("profiles:      %s\n", p.ProfilesDir)
		return nil
	case "install-bundled":
		return installBundled(m)
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		// bare profile name: skill codex → use codex
		if len(args) == 1 {
			if _, err := os.Stat(filepath.Join(p.ProfilesDir, args[0]+".txt")); err == nil {
				return m.UseProfile(args[0])
			}
		}
		return fmt.Errorf("unknown command %q (try skill help)", args[0])
	}
}

func cmdList(m *manager.Manager) error {
	all, err := m.SkillNames()
	if err != nil {
		return err
	}
	agents, err := m.EnabledAgents()
	if err != nil {
		return err
	}
	claude, err := m.EnabledOn(m.P.ClaudeActive)
	if err != nil {
		return err
	}
	aSet := map[string]bool{}
	for _, n := range agents {
		aSet[n] = true
	}
	cSet := map[string]bool{}
	for _, n := range claude {
		cSet[n] = true
	}
	cur, _ := m.CurrentProfile()
	if cur != "" {
		fmt.Printf("profile: %s\n", cur)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tAGENTS\tCLAUDE\tOK")
	for _, name := range all {
		ag := boolMark(aSet[name])
		cl := boolMark(cSet[name])
		ok := " "
		if aSet[name] == cSet[name] {
			ok = "✓"
		} else {
			ok = "!"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, ag, cl, ok)
	}
	_ = w.Flush()
	fmt.Printf("\n%d in warehouse, %d enabled (agents)\n", len(all), len(agents))
	return nil
}

func boolMark(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func cmdTUI(m *manager.Manager) error {
	all, err := m.SkillNames()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("warehouse empty — run: skill sync")
		return nil
	}
	agents, err := m.EnabledAgents()
	if err != nil {
		return err
	}
	aSet := map[string]bool{}
	for _, n := range agents {
		aSet[n] = true
	}
	items := make([]tui.Item, 0, len(all))
	for _, name := range all {
		items = append(items, tui.Item{Name: name, Enabled: aSet[name]})
	}
	enabled, apply, err := tui.Run(items)
	if err != nil {
		return err
	}
	if !apply {
		fmt.Println("cancelled")
		return nil
	}
	if err := m.ApplySet(enabled); err != nil {
		return err
	}
	cur, _ := m.CurrentProfile()
	if cur == "" {
		cur = "default"
	}
	_ = m.SaveProfile(cur, enabled)
	_ = m.SetCurrentProfile(cur)
	sort.Strings(enabled)
	fmt.Printf("applied %d skills → profile %q\n", len(enabled), cur)
	return nil
}

func printSync(res *manager.SyncResult) {
	fmt.Printf("migrated: %s\n", strings.Join(res.Migrated, ", "))
	fmt.Printf("updated:  %s\n", strings.Join(res.Updated, ", "))
	if len(res.Warnings) > 0 {
		fmt.Println("warnings:")
		for _, w := range res.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
	fmt.Printf("enabled:  %d skills\n", len(res.Enabled))
}

func installBundled(m *manager.Manager) error {
	// bundled skills live next to module: skills/<name>
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "skills"))
	}
	wd, _ := os.Getwd()
	candidates = append(candidates,
		filepath.Join(wd, "skills"),
		filepath.Join(wd, "bundled"),
	)
	var srcRoot string
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			srcRoot = c
			break
		}
	}
	if srcRoot == "" {
		return fmt.Errorf("bundled skills directory not found (looked for ./skills)")
	}
	if err := m.P.EnsureDirs(); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		return err
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
			fmt.Printf("skip %s (already in warehouse)\n", e.Name())
			continue
		}
		if err := manager.CopyDir(src, dst); err != nil {
			return err
		}
		installed = append(installed, e.Name())
	}
	fmt.Printf("installed bundled: %s\n", strings.Join(installed, ", "))
	_ = m.EnsureCoreProfile()
	return nil
}

func printHelp() {
	fmt.Print(`skill — manage agent skills (warehouse + dual activity sets)

Usage:
  skill                 Interactive TUI (space toggle, enter apply)
  skill list            List warehouse / enabled (agents + claude)
  skill enable  <n...>  Enable on both .agents/skills and .claude/skills
  skill disable <n...>  Disable on both
  skill sync            Ingest real dirs / fix links into skills-all
  skill doctor          Detect two-level links, drift, real dirs
  skill use <profile>   Apply profile to both activity sets
  skill <profile>       Shorthand for use (if profile exists)
  skill save <profile>  Save current agents enabled set
  skill profiles        List profiles (* = current)
  skill init            Switch to core profile
  skill paths           Show resolved directories
  skill install-bundled Copy repo skills/ into warehouse
  skill help            This help

Env:
  SKILL_MANAGER_HOME    Override ~/.agents
  SKILL_MANAGER_CLAUDE  Override ~/.claude/skills
`)
}
