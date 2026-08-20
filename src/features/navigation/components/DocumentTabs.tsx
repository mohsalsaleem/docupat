import type { Document } from '../../documents';

export function DocumentTabs(props: { document: Document | null; dirty: boolean }) {
  return <div className="flex h-10 items-end border-b border-white/[.07] bg-[var(--surface-1)] px-2">
    {props.document ? <div className="relative flex h-9 min-w-48 max-w-72 items-center gap-2 rounded-t-lg border border-b-0 border-white/[.08] bg-[var(--surface-2)] px-3 text-xs text-zinc-300">
      <span className="text-[var(--accent)]">◆</span>
      <span className="min-w-0 flex-1 truncate">{props.document.title}</span>
      {props.dirty ? <span className="size-1.5 rounded-full bg-zinc-400" title="Unsaved changes" /> : null}
      <span className="absolute inset-x-0 bottom-0 h-px bg-[var(--surface-2)]" />
    </div> : null}
  </div>;
}
