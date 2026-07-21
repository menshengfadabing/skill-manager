# Stack Choice — skill-manager

| 维度 | 选择 | 理由 |
| --- | --- | --- |
| language | Go 1.23 | 单二进制、跨 OS、symlink API 清晰；排除 shell（易卡壳） |
| package_manager | go modules | 官方标准 |
| framework | cobra 风格自研 CLI + bubbletea TUI | 轻量；TUI 用 charmbracelet/bubbletea |
| base_image | N/A（本地 CLI，首期不容器化） | — |
| deploy | go install / 发布 binary | 首期 |

## Decision override

用户已拍板 Go + bubbletea；CLI 为主、companion skill 极薄不互包。

## Toolchain

见同目录选型时 `stack --check-tools`：go / gopls / uv / node 均可用。
