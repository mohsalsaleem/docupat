export function StatusBar(props: { version?: number; selected: number; busy: boolean; message: string; words:number }) {
  return <footer className="flex h-7 items-center justify-between border-t border-white/[.07] bg-[var(--surface-1)] px-3 text-[10px] text-zinc-500">
    <div className="flex min-w-0 items-center gap-3"><span className={`size-1.5 rounded-full ${props.busy ? 'animate-pulse bg-amber-300' : 'bg-emerald-400'}`} /><span className="truncate">{props.message}</span></div>
    <div className="ml-4 flex shrink-0 items-center gap-4"><span>{props.words.toLocaleString()} words · {Math.max(1,Math.ceil(props.words/220))} min read</span><span>{props.selected ? `${props.selected} selected` : 'No selection'}</span><span>Markdown</span>{props.version ? <span>v{props.version}</span> : null}</div>
  </footer>;
}
