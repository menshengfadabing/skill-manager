# 官网（Vite + React + TypeScript）

```bash
cd web
npm install
npm run dev          # 本地预览
npm run build        # 输出到 ../dist/www
npm run deploy:prepare  # 合并 dist/www + dist/artifacts → dist/hosting
```

静态托管部署：

```bash
tcb hosting deploy ../dist/hosting -e menshengfadabing-d3ep6tl006fe480 --yes
```

安装包（二进制 / install 脚本 / skills）放在 `../dist/artifacts/`，与站点构建产物分开，避免被前端构建清空。
