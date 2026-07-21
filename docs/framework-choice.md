# Framework Choice — skill-manager

| 项 | 选择 |
| --- | --- |
| AI 编程框架 | 轻量自研（无 openspec/superpowers）；按 docs/skill-manager-design.md 落地 |
| 测试 | 标准 `go test`，核心 sync/enable 路径单测 |
| 留痕 | `docs/`（design / stack / framework / phases） |
| Git | go mod 无自动 init → 补 git init + 首次 commit |

## 范围

首期：sync / doctor / list / enable / disable / use / save / init / TUI；双侧活动集直连 skills-all。
