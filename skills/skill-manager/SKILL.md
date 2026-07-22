---
name: skill-manager
description: Manage which agent skills are enabled to reduce token waste. Use when the user wants to toggle skills via TUI, sync skill dirs, switch profiles, rollback snapshots, uninstall, or complains about too many skills. Always invoke the `skill` CLI — do not manually create symlinks. Global-only: one warehouse + global work dirs; switch enable sets with profiles (not project vs -g).
---

# skill-manager (companion)

Thin companion for the **`skill` CLI**. Do not re-implement filesystem steps here.

## Model

- **Global only**: warehouse `~/.agents/skills-all`; work dirs under `$HOME` for Claude Code / Codex / Cursor / Qwen Code / Pi (see `docs/tools-paths.md`)
- Switch which skills are on with **profiles** (`skill use <name>`), not project-vs-global scope
- Shared hub: `~/.agents/skills` (Codex/Cursor/Qwen/Pi). Also mirrors tool-specific dirs
- Do **not** manage `~/.cursor/skills-cursor`
- `-g` / `--global` is a no-op (kept for old habits)

## Commands

```bash
skill
skill list
skill create <name>
skill delete <name> [--force]
skill profile
skill use <name>
skill doctor
skill sync [--dry-run] [--yes]
skill init [--yes]
skill log
skill restore <id|initial> [--yes]
skill uninstall [--restore-initial] [--yes]
```

## Rules

1. Prefer CLI over hand-editing skill directories.
2. Never create two-level links (`.claude → .agents/skills → …`).
3. `sync` / `init` are destructive: confirmation or `--yes`; auto-backup under `~/.agents/backups/`（含「用户初始」）.
4. Rollback with `skill log` + `skill restore`；离开本工具用 `skill uninstall`.
