import type { Document, Patch, PatchStatus } from '../../documents/types';
import { headings } from '../../editor/markdown';
import { SectionLabel } from '../../../ui';

interface SidebarProps { documents:Document[]; active:Document|null; draft:string; patches:Patch[]; onOpen:(id:string)=>void; onSelect:(start:number,end:number)=>void }

export function Sidebar(props: SidebarProps) {
  return <aside className="overflow-auto border-r border-white/10 bg-[#101318] p-4">
    <SectionLabel>Documents</SectionLabel>
    {props.documents.map((item) => <button key={item.id} onClick={() => props.onOpen(item.id)} className={`mb-1 w-full rounded-lg px-3 py-2 text-left text-xs ${props.active?.id === item.id ? 'bg-lime-300/10 text-lime-200' : 'text-zinc-400 hover:bg-white/5'}`}>
      {item.title}<span className="float-right text-zinc-600">v{item.version}</span>
    </button>)}
    <SectionLabel spaced>Outline</SectionLabel>
    {headings(props.draft).map((heading) => <button key={`${heading.start}`} onClick={() => props.onSelect(heading.start, heading.end)} style={{ paddingLeft: `${8 + (heading.level - 1) * 12}px` }} className="block w-full truncate rounded px-2 py-1.5 text-left text-xs text-zinc-400 hover:bg-white/5 hover:text-white">{heading.title}</button>)}
    <SectionLabel spaced>Recent patches</SectionLabel>
    {props.patches.slice(0, 8).map((item) => <div key={item.id} className="mb-2 border-l border-white/10 pl-2 text-[11px] text-zinc-500">
      <span className={statusColor(item.status)}>{item.status}</span>{' · '}{item.instruction}
    </div>)}
  </aside>;
}

function statusColor(status: PatchStatus) { return status === 'applied' ? 'text-emerald-400' : status === 'rejected' ? 'text-red-400' : 'text-amber-300'; }
