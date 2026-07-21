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

type flags struct {
	global bool
	yes    bool
	force  bool
	dryRun bool
	args   []string
}

func parseFlags(argv []string) flags {
	f := flags{}
	var rest []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch a {
		case "-g", "--global":
			f.global = true
		case "--yes", "-y":
			f.yes = true
		case "--force":
			f.force = true
		case "--dry-run":
			f.dryRun = true
		case "--help", "-h":
			rest = append(rest, "help")
		default:
			rest = append(rest, a)
		}
	}
	f.args = rest
	return f
}

func run(argv []string) error {
	f := parseFlags(argv)
	if len(f.args) > 0 && (f.args[0] == "help") {
		printHelp()
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	p, err := paths.Resolve(cwd, f.global)
	if err != nil {
		return err
	}
	m := manager.New(p)

	if len(f.args) == 0 {
		return cmdTUI(m)
	}

	switch f.args[0] {
	case "list", "ls":
		return cmdList(m)
	case "sync":
		return cmdSync(m, f)
	case "doctor":
		return cmdDoctor(m)
	case "use":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill use <profile> [-g]")
		}
		return m.UseProfile(f.args[1])
	case "create":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill create <profile> [-g]")
		}
		desc := ""
		if len(f.args) > 2 {
			desc = strings.Join(f.args[2:], " ")
		}
		if err := m.CreateProfile(f.args[1], desc); err != nil {
			return err
		}
		fmt.Printf("created empty profile %q (%s)\n", f.args[1], p.Scope)
		return nil
	case "delete":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill delete <profile> [-g] [--force]")
		}
		if err := m.DeleteProfile(f.args[1], f.force); err != nil {
			return err
		}
		fmt.Printf("deleted profile %q\n", f.args[1])
		return nil
	case "profile":
		return cmdProfile(m)
	case "init":
		return cmdInit(m, f)
	case "help":
		printHelp()
		return nil
	default:
		if len(f.args) == 1 {
			if _, err := os.Stat(filepath.Join(p.ProfilesDir, f.args[0]+".yaml")); err == nil {
				return m.UseProfile(f.args[0])
			}
		}
		return fmt.Errorf("unknown command %q (try skill help)", f.args[0])
	}
}

func cmdList(m *manager.Manager) error {
	fmt.Println(m.P.Summary())
	fmt.Println()
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
	aSet, cSet := toBoolMap(agents), toBoolMap(claude)
	cur, _ := m.CurrentProfile()
	if cur != "" {
		fmt.Printf("profile: %s\n", cur)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tAGENTS\tCLAUDE\tOK")
	for _, name := range all {
		ag, cl := aSet[name], cSet[name]
		ok := "✓"
		if ag != cl {
			ok = "!"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, boolMark(ag), boolMark(cl), ok)
	}
	_ = w.Flush()
	fmt.Printf("\n%d in warehouse, %d enabled (agents) [%s]\n", len(all), len(agents), m.P.Scope)
	return nil
}

func toBoolMap(names []string) map[string]bool {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return m
}

func boolMark(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func cmdDoctor(m *manager.Manager) error {
	fmt.Println(m.P.Summary())
	fmt.Println()
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
}

func cmdProfile(m *manager.Manager) error {
	fmt.Printf("scope=%s profiles=%s\n", m.P.Scope, m.P.ProfilesDir)
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
}

func cmdSync(m *manager.Manager, f flags) error {
	if f.dryRun {
		res, err := m.Sync(manager.SyncOptions{DryRun: true})
		if err != nil {
			return err
		}
		fmt.Printf("[dry-run] scope=%s\n", m.P.Scope)
		fmt.Printf("would migrate: %s\n", strings.Join(res.WouldMigrate, ", "))
		fmt.Printf("would update:  %s\n", strings.Join(res.WouldUpdate, ", "))
		return nil
	}
	prompt := fmt.Sprintf("即将对 [%s] 执行 sync（迁移实体目录、重建软链）。\n仓库: %s\n活动集: .agents/skills + .claude/skills（Qwen 读 .agents，无需镜像）\n会先写入备份到 %s",
		m.P.Scope, m.P.Warehouse, m.P.BackupsDir)
	if err := manager.Confirm(prompt, f.yes); err != nil {
		return err
	}
	res, err := m.Sync(manager.SyncOptions{Yes: f.yes})
	if err != nil {
		return err
	}
	printSync(res)
	return nil
}

func cmdInit(m *manager.Manager, f flags) error {
	cur, _ := m.CurrentProfile()
	prompt := fmt.Sprintf("即将对 [%s] 执行 init：活动集清空为 core（skill-manager + skill-init）。\n当前 profile: %s\n可之后用 skill use <profile> 恢复；备份目录: %s",
		m.P.Scope, cur, m.P.BackupsDir)
	if err := manager.Confirm(prompt, f.yes); err != nil {
		return err
	}
	bak, err := m.BackupSnapshot("init")
	if err != nil {
		return err
	}
	if _, err := m.InstallBundled(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: bundled: %v\n", err)
	}
	if err := m.Init(); err != nil {
		return err
	}
	fmt.Printf("switched to core profile [%s]\nbackup: %s\n", m.P.Scope, bak)
	return nil
}

func cmdTUI(m *manager.Manager) error {
	fmt.Printf("[%s] %s\n", m.P.Scope, m.P.Root)
	all, err := m.SkillNames()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("warehouse 为空 — 先运行: skill sync   （全局则 skill sync -g）")
		return nil
	}
	agents, err := m.EnabledAgents()
	if err != nil {
		return err
	}
	aSet := toBoolMap(agents)
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
		fmt.Println("没有当前 profile：请先 skill create <name> 再 skill use <name>，或再次打开 TUI 前先 use。")
		fmt.Println("本次已应用到活动集，但未写入 profile 文件。")
		return nil
	}
	if err := m.SaveProfile(cur, enabled); err != nil {
		return err
	}
	sort.Strings(enabled)
	fmt.Printf("applied %d skills → profile %q [%s]\n", len(enabled), cur, m.P.Scope)
	return nil
}

func printSync(res *manager.SyncResult) {
	if res.BackupDir != "" {
		fmt.Printf("backup:   %s\n", res.BackupDir)
	}
	if len(res.Bundled) > 0 {
		fmt.Printf("bundled:  %s\n", strings.Join(res.Bundled, ", "))
	}
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

func printHelp() {
	fmt.Print(`skill — Agent Skills 管理（默认项目级；-g 全局）

范围:
  默认          当前 git 仓库：.agents/skills + .claude/skills（CC 例外）
  -g --global   ~/.agents/skills + ~/.claude/skills
  Qwen/Cursor/Codex 读 .agents/skills，不镜像 .qwen
  不维护项目 .codex/skills

命令:
  skill                 交互界面（空格切换，回车应用，写回当前 profile）
  skill list            列出 warehouse（项目∪全局继承）与 agents/claude 启用状态
  skill create <名>     创建空 YAML profile
  skill delete <名>     删除 profile（当前正在用的需 --force）
  skill profile         列出 profile（* 为当前）
  skill use <名>        应用 profile 到本 scope 镜像目录
  skill <名>            若 profile 存在，等同 use
  skill doctor          体检（含路径摘要）
  skill sync            迁移实体、修链、补 bundled（破坏性：需确认或 --yes）
  skill sync --dry-run  只预览，不写盘
  skill init            切到 core 最小集（破坏性：需确认或 --yes）
  skill help            本帮助

标志:
  -g --global   操作全局 scope
  --yes -y      跳过破坏性确认（CI/非 TTY 必须）
  --force       允许删除当前 profile
  --dry-run     仅 sync：预览

说明:
  项目/全局共用 ~/.agents/skills-all；项目只管理活动集 + profiles
  仅 Claude Code 需要 .claude/skills 镜像；Qwen 扫 .agents 即可
  不做同项目按工具分叉活动集

安全:
  sync/init 会先备份到 <scope>/.agents/backups/<时间戳>/
  未加 -g 时不改 ~/.agents 活动集

环境变量:
  SKILL_MANAGER_HOME    覆盖 ~/.agents（全局 warehouse 根）
  SKILL_MANAGER_CLAUDE  覆盖 ~/.claude/skills
`)
}
