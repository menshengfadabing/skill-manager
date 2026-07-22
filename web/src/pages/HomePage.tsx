import { useState } from 'react'
import { Nav } from '../components/Nav'
import { CopyBlock } from '../components/CopyBlock'
import { Topology } from '../components/Topology'
import '../components/Topology.css'
import { HOST } from '../constants'

type Os = 'unix' | 'win'

export function HomePage() {
  const defaultOs: Os = /windows/i.test(navigator.userAgent) ? 'win' : 'unix'
  const [goOs, setGoOs] = useState<Os>(defaultOs)
  const [scriptOs, setScriptOs] = useState<Os>(defaultOs)

  const goUnix = `# 安装 Go 1.23+（以官网包为准：https://go.dev/dl/）
curl -fsSL https://go.dev/dl/go1.23.4.linux-amd64.tar.gz -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
go version`

  const goWin = `# 1) 浏览器打开 https://go.dev/dl/ 下载 Windows MSI 并安装
# 2) 重新打开 PowerShell，确认版本：
go version`

  const src = `git clone https://github.com/menshengfadabing/skill-manager.git
cd skill-manager && go install ./cmd/skill
export SKILL_MANAGER_BUNDLED="$PWD/skills"
skill sync --yes`

  const unixScript = `curl -fsSL ${HOST}/install.sh | bash`
  const winScript = `iwr -UseBasicParsing ${HOST}/install.ps1 -OutFile "$env:TEMP\\sm-install.ps1"; powershell -ExecutionPolicy Bypass -File "$env:TEMP\\sm-install.ps1"`

  return (
    <div className="app-shell">
      <Nav current="home" />
      <main>
        <h1 className="brand"><span>skill-manager</span></h1>

        <section>
          <h2 className="sec-title">😵 技能一堆之后 <span className="chip bad">痛点场</span></h2>
          <Topology
            label="skill 堆积痛点拓扑"
            hubEmoji="🌋"
            hubTitle="100+"
            hubSub="skills 堆积"
            hubKind="pain"
            nodes={[
              {
                kind: 'pain',
                tag: '💸 渐进式披露？',
                body: <>一句 hi，上下文里塞进 <strong>1w+ token</strong> 的 skill 说明——积少成多，token 浪费。</>,
              },
              {
                kind: 'pain',
                tag: '🎯 触发不准',
                body: '模型注意力被摊薄。想用的 skill 没出场，不想用的却抢戏。装得越多，命中越玄学。',
              },
              {
                kind: 'pain',
                tag: '⚔️ skill 冲突',
                body: 'A / B 触发词差不多，规范却相反——模型左右横跳，你也懵。',
              },
              {
                kind: 'pain',
                tag: '🗂️ 管理混乱',
                body: '「我记得 Claude Code 里有这个 skill 啊？怎么只有 Codex 里有？」路径各玩各的。',
              },
            ]}
          />

          <h2 className="sec-title" style={{ marginTop: '2rem' }}>
            ✨ 收成「仓库目录 + 工作目录」 <span className="chip good">解法场</span>
          </h2>
          <Topology
            label="skill-manager 收益拓扑"
            hubEmoji="🧭"
            hubTitle={<>skill<br />manager</>}
            hubSub="按需启用"
            hubKind="win"
            nodes={[
              {
                kind: 'win',
                tag: '🧠 少占上下文',
                body: '只启用当前需要的一小撮，token 留给正事，注意力不再被技能说明书淹没。',
              },
              {
                kind: 'win',
                tag: '🎯 按需触发',
                body: '配置档 / 终端界面明确「这局谁上场」——该出现的出现，不该抢戏的关掉。',
              },
              {
                kind: 'win',
                tag: '🤝 减少互殴',
                body: '同一场景别叠一堆相反规范；工作目录里只留一套说得通的规则。',
              },
              {
                kind: 'win',
                tag: '🔗 一处实体',
                body: '技能文件进仓库目录，各工具工作目录用软链对齐——不再「这个工具有、那个没有」。',
              },
            ]}
          />
        </section>

        <section style={{ marginTop: '2.4rem' }}>
          <h2 className="sec-title">🚀 安装</h2>
          <div className="panel">
            <h3>配置 Go 环境</h3>
            <div className="seg" role="tablist" aria-label="Go 安装系统">
              <button type="button" className={`btn${goOs === 'unix' ? ' active' : ''}`} aria-selected={goOs === 'unix'} onClick={() => setGoOs('unix')}>
                macOS / Linux
              </button>
              <button type="button" className={`btn${goOs === 'win' ? ' active' : ''}`} aria-selected={goOs === 'win'} onClick={() => setGoOs('win')}>
                Windows
              </button>
            </div>
            {goOs === 'unix' ? (
              <CopyBlock id="c-go-unix">{goUnix}</CopyBlock>
            ) : (
              <CopyBlock id="c-go-win">{goWin}</CopyBlock>
            )}

            <h3>方式一：源码安装</h3>
            <CopyBlock id="c-src">{src}</CopyBlock>

            <h3>方式二：脚本安装</h3>
            <div className="seg" role="tablist" aria-label="脚本安装系统">
              <button type="button" className={`btn${scriptOs === 'unix' ? ' active' : ''}`} aria-selected={scriptOs === 'unix'} onClick={() => setScriptOs('unix')}>
                macOS / Linux
              </button>
              <button type="button" className={`btn${scriptOs === 'win' ? ' active' : ''}`} aria-selected={scriptOs === 'win'} onClick={() => setScriptOs('win')}>
                Windows
              </button>
            </div>
            {scriptOs === 'unix' ? (
              <CopyBlock id="c-unix">{unixScript}</CopyBlock>
            ) : (
              <CopyBlock id="c-win">{winScript}</CopyBlock>
            )}
          </div>
        </section>

        <footer className="page-foot" />
      </main>
    </div>
  )
}
