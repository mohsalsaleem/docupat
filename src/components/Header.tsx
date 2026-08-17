interface HeaderProps { preview:boolean; dirty:boolean; busy:boolean; onTogglePreview:()=>void; onSave:()=>void }

export function Header(props: HeaderProps) {
  return <header className="flex h-16 items-center justify-between border-b border-white/10 px-5">
    <div className="flex items-center gap-3">
      <div className="grid size-9 place-items-center rounded-lg bg-lime-300 font-black text-black">D/</div>
      <div><b>DocPatch</b><div className="text-[10px] uppercase tracking-[.18em] text-zinc-500">precision document editing</div></div>
    </div>
    <div className="flex gap-2">
      <button onClick={props.onTogglePreview} className="rounded-lg border border-white/10 px-3 py-2 text-xs hover:bg-white/5">{props.preview ? 'Editor' : 'Diagrams'}</button>
      <button disabled={props.busy || !props.dirty} onClick={props.onSave} className="rounded-lg bg-zinc-100 px-4 py-2 text-xs font-bold text-black disabled:opacity-30">Save</button>
    </div>
  </header>;
}
