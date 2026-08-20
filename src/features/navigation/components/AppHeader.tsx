import { Button } from '../../../ui';

interface HeaderProps { preview:boolean; dirty:boolean; busy:boolean; onTogglePreview:()=>void; onSave:()=>void }

export function AppHeader(props: HeaderProps) {
  return <header className="flex h-12 items-center justify-between border-b border-white/[.07] bg-[var(--surface-1)] px-3">
    <div className="flex items-center gap-3">
      <div className="grid size-7 place-items-center rounded-lg bg-[var(--accent)] text-xs font-black text-[#11130f] shadow-[0_0_18px_var(--accent-glow)]">S</div>
      <div className="flex items-baseline gap-2"><b className="text-sm font-semibold tracking-tight">Stellarity</b><span className="text-[10px] text-zinc-600">technical workspace</span></div>
    </div>
    <div className="flex gap-2">
      <Button onClick={props.onTogglePreview}>{props.preview ? 'Editor' : 'Diagrams'}</Button>
      <Button variant="secondary" disabled={props.busy || !props.dirty} onClick={props.onSave}>Save</Button>
    </div>
  </header>;
}
