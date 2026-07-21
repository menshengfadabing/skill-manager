---
name: skill-manager
description: Manage which agent skills are enabled to reduce token waste. Use when the user wants to enable/disable skills, run skill TUI, sync skill directories, switch profiles (codex/default/core), or complains that too many skills dilute attention. Always invoke the `skill` CLI — do not manually create symlinks.
---

# skill-manager (companion)

Thin companion for the **`skill` CLI**. Do not re-implement filesystem or symlink steps here.

## When to use

- User asks to enable/disable/list skills
- Too many skills / token waste / attention dilution
- Switch scene profile (`codex`, `default`, `core`)
- After installing skills via `npx skills` → remind `skill sync`

## Commands (run in shell)

```bash
skill              # TUI: space toggle, enter apply (both .agents + .claude)
skill list
skill enable <name...>
skill disable <name...>
skill sync         # ingest real dirs / fix two-level links into skills-all
skill doctor
skill use <profile>
skill save <profile>
skill init         # core-only profile
skill paths
```

## Rules

1. Prefer `skill` CLI over editing `~/.agents/skills` or `~/.claude/skills` by hand.
2. Never create `.claude/skills/X → .agents/skills/X` (two-level). CLI links both sides **directly** to `~/.agents/skills-all/X`.
3. After external installs land in an activity set, run `skill sync`.
