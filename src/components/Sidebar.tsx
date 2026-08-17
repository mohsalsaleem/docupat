import type { Document, Patch } from '../api';
import { headings } from '../diff';

interface SidebarProps { documents:Document[]; active:Document|null; draft:string; patches:Patch[]; onOpen:(id:string)=>void; onSelect:(start:number,end:number)=>void }

export function Sidebar(props: SidebarProps) {
  return <aside className="overflow-auto border-r border-white/10 bg-[#101318] p-4">
    <Label>Documents</Label>
    {props.documents.map((item) => <button key={item.id} onClick={() => props.onOpen(item.id)} className={`mb-1 w-full rounded-lg px-3 py-2 text-left text-xs ${props.active?.id === item.id ? 'bg-lime-300/10 text-lime-200' : 'text-zinc-400 hover:bg-white/5'}`}>
      {item.title}<span className="float-right text-zinc-600">v{item.version}</span>
    </button>)}
    <Label spaced>Outline</Label>
    {headings(props.draft).map((heading) => <button key={`${heading.start}`} onClick={() => props.onSelect(heading.start, heading.end)} style={{ paddingLeft: `${8 + (heading.level - 1) * 12}px` }} className="block w-full truncate rounded px-2 py-1.5 text-left text-xs text-zinc-400 hover:bg-white/5 hover:text-white">{heading.title}</button>)}
    <Label spaced>Recent patches</Label>
    {props.patches.slice(0, 8).map((item) => <div key={item.id} className="mb-2 border-l border-white/10 pl-2 text-[11px] text-zinc-500">
      <span className={statusColor(item.status)}>{item.status}</span>{' · '}{item.instruction}
    </div>)}
  </aside>;
}

function Label(props:{children:unknown;spaced?:boolean}) { return <div className={`${props.spaced ? 'mt-7 ' : ''}mb-3 text-[10px] font-bold uppercase tracking-[.18em] text-zinc-600`}>{props.children}</div>; }
function statusColor(status:Patch['status']) { return status === 'applied' ? 'text-emerald-400' : status === 'rejected' ? 'text-red-400' : 'text-amber-300'; }
