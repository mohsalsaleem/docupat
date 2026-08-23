import type { ContextAssessment, ContextItem } from '../../documents/types';

export function ContextSummary({ items, assessment }: { items: ContextItem[]; assessment?: ContextAssessment }) {
  return <section className="mb-4 rounded-lg border border-white/[.07] bg-black/20 p-3">
    {assessment?.level ? <div className="mb-3 flex items-center justify-between border-b border-white/[.06] pb-3">
      <div><div className="text-xs font-medium text-zinc-200">{confidenceTitle(assessment.level)}</div><div className="mt-1 text-[10px] text-zinc-500">Based on your selection and related parts of the workspace</div></div>
      <span className={`rounded-full border px-2 py-1 text-[10px] font-semibold ${confidenceColor(assessment.level)}`}>{assessment.level === 'low' ? 'Review carefully' : 'Well supported'}</span>
    </div> : null}
    <div className="mb-2 flex items-center justify-between text-[10px] font-semibold uppercase tracking-[.14em] text-zinc-500">
      <span>Used for this suggestion</span>
      <span>{items.length} related {items.length === 1 ? 'section' : 'sections'}</span>
    </div>
    {items.length ? <div className="flex flex-wrap gap-2">
      {items.map((item, index) => <span key={`${item.kind}-${item.title}-${index}`} title={item.content} className="rounded-md border border-white/[.08] bg-white/[.04] px-2 py-1 text-[11px] text-zinc-300">
        <span className="mr-1.5 text-[var(--accent)]">{symbol(item.kind)}</span>{item.title}
        <span className="ml-1.5 text-zinc-500">{item.documentTitle}</span>
      </span>)}
    </div> : <p className="text-xs text-zinc-600">Only your selection and its section were needed.</p>}
  </section>;
}

function confidenceTitle(level: ContextAssessment['level']) {
  if (level === 'high') return 'This suggestion has strong supporting context';
  if (level === 'medium') return 'This suggestion has useful supporting context';
  return 'This suggestion has limited supporting context';
}

function confidenceColor(level: ContextAssessment['level']) {
  if (level === 'high') return 'border-lime-300/20 bg-lime-300/10 text-lime-200';
  if (level === 'medium') return 'border-amber-300/20 bg-amber-300/10 text-amber-200';
  return 'border-red-300/20 bg-red-300/10 text-red-200';
}

function symbol(kind: ContextItem['kind']) {
  return kind === 'ancestor' ? '↳' : kind === 'backlink' ? '←' : kind === 'semantic' ? '≈' : '→';
}
