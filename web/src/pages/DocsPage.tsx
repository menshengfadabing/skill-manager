import { Nav } from '../components/Nav'
import { PATHS } from '../constants'
import './DocsPage.css'

const NAV = [
  { id: 'paths', label: '目录约定' },
  { id: 'commands', label: '常用命令' },
  { id: 'snapshot', label: '快照与回滚' },
  { id: 'uninstall', label: '卸载清理' },
  { id: 'companions', label: '配套 skill' },
  { id: 'flags', label: '标志与环境变量' },
] as const

export function DocsPage() {
  return (
    <div className="app-shell">
      <Nav current="docs" />
      <div className="docs-layout">
        <aside className="docs-side" aria-label="文档目录">
          <p className="side-title">目录</p>
          <nav className="side-nav">
            {NAV.map((item) => (
              <a key={item.id} className="side-link" href={`#${item.id}`}>
                {item.label}
              </a>
            ))}
          </nav>
        </aside>

        <main className="docs-main">
          <h1 className="docs-title">文档</h1>

          <section id="paths" className="docs-sec">
            <h2>目录约定</h2>
            <p className="lead">首批支持：Claude Code / Codex / Cursor / Qwen Code / Pi。详细表见仓库 <code>docs/tools-paths.md</code>。</p>
            <div className="glossary panel">
              <table>
                <thead>
                  <tr><th>叫法</th><th>路径</th><th>干什么</th></tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong>仓库目录</strong></td>
                    <td><code>{PATHS.warehouse}</code></td>
                    <td>所有 skill 实体只存这一份</td>
                  </tr>
                  <tr>
                    <td><strong>共享工作目录</strong></td>
                    <td><code>{PATHS.workGlobal}</code></td>
                    <td>Codex / Cursor / Qwen / Pi 共用枢纽</td>
                  </tr>
                  <tr>
                    <td><strong>Claude</strong></td>
                    <td><code>{PATHS.workClaudeGlobal}</code></td>
                    <td>Claude Code 专用</td>
                  </tr>
                  <tr>
                    <td><strong>Cursor</strong></td>
                    <td><code>{PATHS.workCursorGlobal}</code></td>
                    <td>不管内置 <code>skills-cursor</code></td>
                  </tr>
                  <tr>
                    <td><strong>Codex 遗留</strong></td>
                    <td><code>{PATHS.workCodexGlobal}</code></td>
                    <td>与共享枢纽一并扫描</td>
                  </tr>
                  <tr>
                    <td><strong>Qwen / Pi</strong></td>
                    <td><code>{PATHS.workQwenGlobal}</code>、<code>{PATHS.workPiGlobal}</code></td>
                    <td>专用目录 + 共享枢纽双写</td>
                  </tr>
                  <tr>
                    <td><strong>配置档</strong></td>
                    <td><code>{PATHS.profiles}*.yaml</code></td>
                    <td>记录启用清单；用 profile 切换「开哪些」</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p className="lead">只维护<strong>全局一套</strong>。不同场景用 <code>skill use &lt;配置档&gt;</code> 切换启用集，不再区分项目级 / <code>-g</code>。</p>
          </section>

          <section id="commands" className="docs-sec">
            <h2>常用命令</h2>
            <div className="cmd">
              <p className="name">skill</p>
              <p>打开交互界面：空格切换启用/禁用，回车应用并写回<strong>当前配置档</strong>。</p>
            </div>
            <div className="cmd">
              <p className="name">skill list</p>
              <p>列出仓库目录里有哪些 skill，以及当前工作目录（含 Claude）的启用状态。</p>
            </div>
            <div className="cmd">
              <p className="name">skill init [--yes]</p>
              <p>把工作目录切到 core 最小集（破坏性）。会先留下快照。非交互请加 <code>--yes</code>。</p>
            </div>
            <div className="cmd">
              <p className="name">skill sync [--dry-run] [--yes]</p>
              <p>扫描工作目录：实体迁入仓库目录，修好软链，补上配套 skill。可用 <code>--dry-run</code> 只预览。</p>
            </div>
            <div className="cmd">
              <p className="name">skill create / delete / profile / use</p>
              <p>管理配置档：创建、删除、列出、应用到工作目录。<code>skill &lt;名&gt;</code> 命中配置档时等同 <code>use</code>。</p>
            </div>
            <div className="cmd">
              <p className="name">skill doctor</p>
              <p>体检：路径是否齐、软链是否指向仓库目录。</p>
            </div>
          </section>

          <section id="snapshot" className="docs-sec">
            <h2>快照与回滚</h2>
            <p className="lead">类似 git：先看历史，再回到某一版。首次执行 <code>sync</code> / <code>init</code> 前，会自动打一份带 <code>*</code> 的<strong>用户初始</strong>快照。</p>
            <div className="cmd">
              <p className="name">skill log</p>
              <p>列出快照。<code>*</code> 标记 = 用户初始（任何破坏性操作之前）。</p>
            </div>
            <div className="cmd">
              <p className="name">skill restore &lt;id|initial&gt; [--yes]</p>
              <p>把工作目录和配置档恢复到指定快照。可用 <code>initial</code> 指向用户初始。恢复前会再打一份 <code>pre-restore</code> 备份。</p>
            </div>
            <pre>{`skill log
skill restore initial --yes
skill restore 20260722-103015-sync --yes`}</pre>
          </section>

          <section id="uninstall" className="docs-sec">
            <h2>卸载清理</h2>
            <div className="cmd">
              <p className="name">skill uninstall [--restore-initial] [--yes]</p>
              <p>清理本工具留下的配置档目录、备份目录，以及工作目录 / 仓库目录里的 <code>skill-manager</code>、<code>skill-init</code>。不会删你其它 skill 实体。</p>
            </div>
            <pre>{`# 先回到装工具前的工作目录，再清痕迹
skill uninstall --restore-initial --yes

# 只清痕迹，不恢复
skill uninstall --yes`}</pre>
          </section>

          <section id="companions" className="docs-sec">
            <h2>配套 skill</h2>
            <div className="cmd">
              <p className="name">skill-manager</p>
              <p>给 Agent 看的薄说明书：提醒用 <code>skill</code> CLI 管理启用集，不要手改软链。覆盖 list / sync / init / 配置档 / 交互界面等用法。</p>
            </div>
            <div className="cmd">
              <p className="name">skill-init</p>
              <p>访谈式「帮项目挑最小 skill 集」：问清项目类型、用哪些 Agent、哪些必须开/必须关，然后通过 <code>skill create</code> / <code>use</code> / 交互界面落地，而不是直接改目录。</p>
            </div>
          </section>

          <section id="flags" className="docs-sec">
            <h2>标志与环境变量</h2>
            <table className="flags">
              <thead><tr><th>标志</th><th>作用</th></tr></thead>
              <tbody>
                <tr><td><code>-g</code> / <code>--global</code></td><td>已废弃（可省略，行为始终为全局）</td></tr>
                <tr><td><code>--yes</code> / <code>-y</code></td><td>跳过破坏性确认</td></tr>
                <tr><td><code>--force</code></td><td>强删当前配置档等</td></tr>
                <tr><td><code>--dry-run</code></td><td>仅 sync：预览</td></tr>
                <tr><td><code>--restore-initial</code></td><td>仅 uninstall：先恢复用户初始</td></tr>
              </tbody>
            </table>
            <table className="flags" style={{ marginTop: '1rem' }}>
              <thead><tr><th>变量</th><th>含义</th></tr></thead>
              <tbody>
                <tr><td><code>SKILL_MANAGER_HOME</code></td><td>覆盖 <code>~/.agents</code></td></tr>
                <tr><td><code>SKILL_MANAGER_CLAUDE</code></td><td>覆盖全局 Claude 工作目录</td></tr>
                <tr><td><code>SKILL_MANAGER_BUNDLED</code></td><td>配套 skill 目录</td></tr>
              </tbody>
            </table>
          </section>
        </main>
      </div>
    </div>
  )
}
