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
			// Deprecated: skill-manager is global-only. Kept as no-op for old habits.
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
	// Always global layout; -g is ignored (compat).
	p, err := paths.Resolve(cwd, true)
	if err != nil {
		return err
	}
	if f.global {
		fmt.Fprintln(os.Stderr, "提示: 已改为仅维护全局一套；-g/--global 可省略。")
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
			return fmt.Errorf("usage: skill use <profile>")
		}
		return cmdUse(m, f.args[1])
	case "create":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill create <profile>")
		}
		desc := ""
		if len(f.args) > 2 {
			desc = strings.Join(f.args[2:], " ")
		}
		if err := m.CreateProfile(f.args[1], desc); err != nil {
			return err
		}
		fmt.Printf("created empty profile %q\n提示: 空配置档 use 后会清空启用集；用 skill 打开 TUI 勾选，或编辑 ~/.agents/profiles/%s.yaml\n", f.args[1], f.args[1])
		return nil
	case "delete":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill delete <profile> [--force]")
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
	case "log":
		return cmdLog(m)
	case "restore":
		if len(f.args) < 2 {
			return fmt.Errorf("usage: skill restore <快照id|initial> [--yes]")
		}
		return cmdRestore(m, f, f.args[1])
	case "uninstall":
		return cmdUninstall(m, f)
	case "help":
		printHelp()
		return nil
	default:
		if len(f.args) == 1 {
			if _, err := os.Stat(filepath.Join(p.ProfilesDir, f.args[0]+".yaml")); err == nil {
				return cmdUse(m, f.args[0])
			}
		}
		return fmt.Errorf("unknown command %q (try skill help)", f.args[0])
	}
}

func cmdUse(m *manager.Manager, name string) error {
	if err := m.UseProfile(name); err != nil {
		names, _ := m.ListProfiles()
		if len(names) > 0 {
			return fmt.Errorf("%w\n已有配置档: %s", err, strings.Join(names, ", "))
		}
		return err
	}
	skills, _ := m.ReadProfile(name)
	fmt.Printf("applied profile %q (%d skills)\n", name, len(skills))
	if len(skills) == 0 {
		fmt.Println("提示: 该配置档技能列表为空，各工具工作目录已清空启用项。用 skill 打开 TUI 勾选后会写回当前配置档。")
	}
	return nil
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
	fmt.Printf("\n%d in warehouse, %d enabled (agents)\n", len(all), len(agents))
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
	fmt.Printf("profiles=%s\n", m.P.ProfilesDir)
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
		fmt.Printf("[dry-run]\n")
		fmt.Printf("would migrate: %s\n", strings.Join(res.WouldMigrate, ", "))
		fmt.Printf("would update:  %s\n", strings.Join(res.WouldUpdate, ", "))
		return nil
	}
	prompt := fmt.Sprintf("即将执行 sync（扫描全局工作目录 + 当前仓库残留目录、迁入仓库、只写回全局启用集）。\n支持: Claude Code / Codex / Cursor / Qwen Code / Pi\n仓库: %s\n备份: %s",
		m.P.Warehouse, m.P.BackupsDir)
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
	prompt := fmt.Sprintf("即将执行 init：全局各工具工作目录清空为 core（skill-manager + skill-init）。\n当前配置档: %s\n备份目录: %s",
		cur, m.P.BackupsDir)
	if err := manager.Confirm(prompt, f.yes); err != nil {
		return err
	}
	if _, err := m.EnsureInitialSnapshot(); err != nil {
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
	fmt.Printf("switched to core profile\nbackup: %s\n", bak)
	return nil
}

func cmdTUI(m *manager.Manager) error {
	fmt.Printf("global %s\n", m.P.Home)
	all, err := m.SkillNames()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("warehouse 为空 — 先运行: skill sync")
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
	fmt.Printf("applied %d skills → profile %q\n", len(enabled), cur)
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

func cmdLog(m *manager.Manager) error {
	snaps, err := m.ListSnapshots()
	if err != nil {
		return err
	}
	if len(snaps) == 0 {
		fmt.Println("暂无快照。首次 skill sync / skill init 前会自动留下「用户初始」。")
		return nil
	}
	fmt.Printf("快照目录: %s\n\n", m.P.BackupsDir)
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "标记\tID\t时间\t说明")
	for _, s := range snaps {
		mark := " "
		note := s.Label
		if s.IsInitial {
			mark = "*"
			if note == "" || note == "initial" {
				note = "用户初始（破坏性操作前）"
			} else {
				note = "用户初始 · " + note
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", mark, s.ID, s.Created.Format("2006-01-02 15:04:05"), note)
	}
	_ = w.Flush()
	fmt.Println("\n恢复: skill restore <ID|initial> [--yes]")
	return nil
}

func cmdRestore(m *manager.Manager, f flags, ref string) error {
	id, err := m.ResolveSnapshotID(ref)
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf("即将把全局工作目录/配置档恢复为快照 %q。\n会先再打一份 pre-restore 备份。\n备份根: %s",
		id, m.P.BackupsDir)
	if err := manager.Confirm(prompt, f.yes); err != nil {
		return err
	}
	if err := m.RestoreSnapshot(id); err != nil {
		return err
	}
	fmt.Printf("已恢复快照 %s\n", id)
	return nil
}

func cmdUninstall(m *manager.Manager, f flags) error {
	restore := false
	for _, a := range f.args[1:] {
		if a == "--restore-initial" || a == "--to-initial" {
			restore = true
		}
	}
	prompt := fmt.Sprintf("即将卸载 skill-manager 痕迹。\n会移除: 配置档、备份、各全局工作目录里的 skill-manager/skill-init，以及仓库中的这两个配套 skill。\n不会删除你其它 skill 实体。\n备份根: %s",
		m.P.BackupsDir)
	if restore {
		prompt += "\n并先恢复「用户初始」快照。"
	} else {
		prompt += "\n若要先回到装之前状态，请加 --restore-initial。"
	}
	if err := manager.Confirm(prompt, f.yes); err != nil {
		return err
	}
	if err := m.Uninstall(manager.UninstallOptions{RestoreInitial: restore, Yes: f.yes}); err != nil {
		return err
	}
	fmt.Println("已卸载 skill-manager 痕迹")
	return nil
}

func printHelp() {
	fmt.Print(`skill — Agent Skills 管理（仅全局一套，用配置档切换启用集）

支持工具: Claude Code / Codex / Cursor / Qwen Code / Pi

模型:
  仓库目录    ~/.agents/skills-all
  工作目录    仅用户主目录下各工具路径（见 docs/tools-paths.md）
  配置档      ~/.agents/profiles/*.yaml   ← 用不同 profile 切换「开哪些」
  不再维护    项目级启用集 / -g 双 scope（避免全局+项目双倍注入）

命令:
  skill                 交互界面（空格切换，回车应用，写回当前配置档）
  skill list            列出仓库与启用状态
  skill create <名>     创建空配置档
  skill delete <名>     删除配置档（当前正在用的需 --force）
  skill profile         列出配置档（* 为当前）
  skill use <名>        应用配置档到全部全局工作目录
  skill <名>            若配置档存在，等同 use
  skill doctor          体检
  skill sync            迁入仓库并重建全局软链；若在 git 仓库内会顺带摄入项目残留
  skill sync --dry-run  只预览
  skill init            全局切到 core 最小集
  skill log             列出快照（* = 用户初始）
  skill restore <id>    恢复快照（initial = 用户初始）
  skill uninstall       清理本工具痕迹；可选 --restore-initial
  skill help            本帮助

标志:
  --yes -y             跳过确认
  --force              允许删除当前配置档
  --dry-run            仅 sync：预览
  --restore-initial    仅 uninstall：先恢复用户初始
  -g --global          已废弃（可省略，行为始终为全局）

环境变量:
  SKILL_MANAGER_HOME    覆盖 ~/.agents
  SKILL_MANAGER_CLAUDE  覆盖 Claude 工作目录
  SKILL_MANAGER_BUNDLED 配套 skill 目录
`)
}
