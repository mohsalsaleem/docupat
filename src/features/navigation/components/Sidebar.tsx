import type { Document, Patch, PatchStatus } from '../../documents/types';
import { headings } from '../../editor/markdown';
import { SectionLabel } from '../../../ui';
import type { NavigationView } from '../types';

interface SidebarProps { view:NavigationView; documents:Document[]; active:Document|null; draft:string; patches:Patch[]; onOpen:(id:string)=>void; onSelect:(start:number,end:number)=>void; onNew:()=>void }

export function Sidebar(props: SidebarProps) {
  return <aside className="min-w-0 overflow-auto border-r border-white/[.07] bg-[var(--surface-1)] px-3 py-4">
    {props.view === 'files' ? <FilesView {...props} /> : null}
    {props.view === 'outline' ? <OutlineView {...props} /> : null}
    {props.view === 'changes' ? <ChangesView {...props} /> : null}
  </aside>;
}

function FilesView(props: SidebarProps) {
  return <><div className="flex items-center justify-between"><SectionLabel>Workspace</SectionLabel><button onClick={props.onNew} className="-mt-1 rounded px-2 py-1 text-lg leading-none text-zinc-500 hover:bg-white/5 hover:text-lime-200" aria-label="New document">+</button></div>
    {props.documents.map((item) => <button key={item.id} onClick={() => props.onOpen(item.id)} className={`mb-1 w-full rounded-lg px-3 py-2 text-left text-xs ${props.active?.id === item.id ? 'bg-lime-300/10 text-lime-200' : 'text-zinc-400 hover:bg-white/5'}`}>
      <span className="mr-2 text-zinc-600">◇</span>{item.title}<span className="float-right text-zinc-600">v{item.version}</span>
    </button>)}
    <div className="mt-6 rounded-lg border border-dashed border-white/[.08] px-3 py-4 text-center text-[11px] leading-5 text-zinc-600">More documents and folders will live here.</div>
  </>;
}

function OutlineView(props: SidebarProps) {
  return <><SectionLabel>Outline</SectionLabel>
    {headings(props.draft).map((heading) => <button key={`${heading.start}`} onClick={() => props.onSelect(heading.start, heading.end)} style={{ paddingLeft: `${8 + (heading.level - 1) * 12}px` }} className="block w-full truncate rounded px-2 py-1.5 text-left text-xs text-zinc-400 hover:bg-white/5 hover:text-white">{heading.title}</button>)}
  </>;
}

function ChangesView(props: SidebarProps) {
  return <><SectionLabel>Changes</SectionLabel>
    {props.patches.slice(0, 8).map((item) => <div key={item.id} className="mb-2 border-l border-white/10 pl-2 text-[11px] text-zinc-500">
      <span className={statusColor(item.status)}>{item.status}</span>{' · '}{item.instruction}
    </div>)}
    {!props.patches.length ? <div className="py-8 text-center text-[11px] text-zinc-600">No AI-assisted changes yet.</div> : null}
  </>;
}

function statusColor(status: PatchStatus) { return status === 'applied' ? 'text-emerald-400' : status === 'rejected' ? 'text-red-400' : 'text-amber-300'; }
