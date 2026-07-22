import type { ReactNode } from 'react'

type Node = {
  tag: string
  body: ReactNode
  kind: 'pain' | 'win'
}

type Props = {
  hubEmoji: string
  hubTitle: ReactNode
  hubSub: string
  hubKind: 'pain' | 'win'
  nodes: [Node, Node, Node, Node]
  label: string
}

export function Topology({ hubEmoji, hubTitle, hubSub, hubKind, nodes, label }: Props) {
  const [tl, tr, bl, br] = nodes
  return (
    <div className={`topo topo-${hubKind}`} aria-label={label}>
      <article className={`node ${tl.kind}`}>
        <div className="tag">{tl.tag}</div>
        <p>{tl.body}</p>
      </article>
      <article className={`node ${tr.kind}`}>
        <div className="tag">{tr.tag}</div>
        <p>{tr.body}</p>
      </article>
      <div className={`hub ${hubKind === 'win' ? 'good' : ''}`}>
        <div className="emoji">{hubEmoji}</div>
        <div className="big">{hubTitle}</div>
        <div className="sub">{hubSub}</div>
      </div>
      <article className={`node ${bl.kind}`}>
        <div className="tag">{bl.tag}</div>
        <p>{bl.body}</p>
      </article>
      <article className={`node ${br.kind}`}>
        <div className="tag">{br.tag}</div>
        <p>{br.body}</p>
      </article>
    </div>
  )
}
