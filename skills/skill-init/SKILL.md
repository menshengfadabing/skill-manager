---
name: skill-init
description: Interview the user about the project, then recommend a minimal skill set. Use when the user runs skill init, wants a clean skill activity set, or asks which skills this project needs. After agreement, enable via `skill enable` / `skill use` — do not edit skill directories manually.
---

# skill-init

## Goal

Produce a **small** enabled skill list for the current project. Prefer 5–10 skills max.

## Interview (ask briefly)

1. What kind of project? (web, CLI, infra, data, docs-only, mixed)
2. Which agents? (Cursor, Claude Code, Codex, …)
3. Must-have workflows? (deploy, TDD, PRD/issues, Cloudflare, Obsidian, …)
4. Any skills that must stay off? (heavy / rare)

## Recommend

- List proposed skill names from the warehouse (`skill list`).
- Explain one line each why.
- Exclude niche skills unless the user confirmed the domain.

## Apply

After the user confirms:

```bash
skill enable <name...>
# or
skill save <project-profile> && skill use <project-profile>
```

If starting from a wipe:

```bash
skill init          # core only (skill-manager + skill-init)
# then enable the agreed set
```

Do not copy or symlink skill directories yourself.
