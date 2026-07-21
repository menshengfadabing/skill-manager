---
name: skill-manager
description: Manage which agent skills are enabled to reduce token waste. Use when the user wants to toggle skills via TUI, sync skill dirs, switch profiles, or complains about too many skills. Always invoke the `skill` CLI — do not manually create symlinks. Default scope is the current git project; use -g for global.
---

# skill-manager (companion)

Thin companion for the **`skill` CLI**. Do not re-implement filesystem steps here.

## Scope

- Default: **current git repo** activity + profiles; skill bodies in shared `~/.agents/skills-all`
- Global activity: `-g` / `--global`
- Qwen / Cursor / Codex read `.agents/skills` — **no** `.qwen` mirror
- Do not use project `.codex/skills` as the managed path
- No per-project `skills-all`

## Commands

```bash
skill                 # TUI
skill list [-g]
skill create <name> [-g]
skill delete <name> [-g] [--force]
skill profile [-g]
skill use <name> [-g]
skill doctor [-g]
skill sync [--dry-run] [--yes] [-g]
skill init [--yes] [-g]
```

## Rules

1. Prefer CLI over hand-editing skill directories.
2. Never create two-level links (`.claude → .agents/skills → …`).
3. `sync` / `init` are destructive: confirmation or `--yes`; auto-backup under `.agents/backups/`.
