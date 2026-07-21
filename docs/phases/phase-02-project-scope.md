# phase-02 — 项目默认 scope + 安全重构

## Done

- `paths.Resolve(cwd, global)`：默认 git root 项目；`-g` 全局
- 镜像目标：`.agents` + `.claude` + `.qwen`（不管理 `.codex`）
- 命令精简：create / delete / profile / use / list / doctor / sync / init / TUI
- Profile 改为 YAML；旧 txt 自动迁移
- sync/init：backup + 确认/`--yes`；sync `--dry-run`；bundled 并入 sync
- README 中文完整说明；单测覆盖项目不碰全局

## Verify

```bash
go test ./...
skill list          # scope=project
skill sync --dry-run
skill -g list       # scope=global
```
