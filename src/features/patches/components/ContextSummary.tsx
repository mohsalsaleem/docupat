import type { ContextItem } from '../../documents/types';

export function ContextSummary({ items }: { items: ContextItem[] }) {
  const characters = items.reduce((total, item) => total + item.content.length, 0);
  const estimatedTokens = Math.ceil(characters / 4);
  return <section className="mb-4 rounded-lg border border-white/[.07] bg-black/20 p-3">
    <div className="mb-2 flex items-center justify-between text-[10px] font-semibold uppercase tracking-[.14em] text-zinc-500">
      <span>Compiled context</span>
      <span>{items.length} {items.length === 1 ? 'source' : 'sources'} · ~{estimatedTokens} tokens</span>
    </div>
    {items.length ? <div className="flex flex-wrap gap-2">
      {items.map((item, index) => <span key={`${item.kind}-${item.title}-${index}`} title={item.content} className="rounded-md border border-white/[.08] bg-white/[.04] px-2 py-1 text-[11px] text-zinc-300">
        <span className="mr-1.5 text-[var(--accent)]">{symbol(item.kind)}</span>{item.title}
        <span className="ml-1.5 text-zinc-500">{item.documentTitle}</span>
        <span className="ml-1.5 text-zinc-600">{item.kind}</span>
        {item.score ? <span className="ml-1 text-zinc-600">{Math.round(item.score * 100)}%</span> : null}
      </span>)}
    </div> : <p className="text-xs text-zinc-600">No linked or structural context was needed.</p>}
  </section>;
}

function symbol(kind: ContextItem['kind']) {
  return kind === 'ancestor' ? '↳' : kind === 'backlink' ? '←' : kind === 'semantic' ? '≈' : '→';
}
