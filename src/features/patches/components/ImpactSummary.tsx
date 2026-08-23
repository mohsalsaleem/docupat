import type { ImpactFinding } from '../../documents/types';

export function ImpactSummary({ items }: { items: ImpactFinding[] }) {
  if (!items.length) return null;
  return <section className="mb-4 rounded-lg border border-amber-300/10 bg-amber-300/[.04] p-3">
    <div className="mb-2 text-[10px] font-semibold uppercase tracking-[.14em] text-amber-200/70">You may also want to review</div>
    <div className="space-y-2">{items.map((item) => <div key={item.sectionId} className="flex items-start justify-between gap-3 text-xs">
      <div><span className="text-zinc-200">{item.title}</span><span className="ml-2 text-zinc-500">{item.documentTitle}</span><div className="mt-0.5 text-[10px] text-zinc-500">{item.reason}</div></div>
    </div>)}</div>
  </section>;
}
