import { useState } from 'octane';
import type { Document } from '../../documents';

export function SearchPalette({documents,onOpen,onClose}:{documents:Document[];onOpen:(id:string)=>void;onClose:()=>void}) {
  const [query,setQuery]=useState('');
  const normalized=query.trim().toLowerCase();
  const matches=documents.filter((item)=>!normalized||`${item.title} ${item.excerpt??''}`.toLowerCase().includes(normalized)).slice(0,10);
  return <div className="fixed inset-0 z-40 flex justify-center bg-black/50 px-6 pt-[12vh] backdrop-blur-sm" role="dialog" aria-modal="true" aria-label="Search documents" onClick={onClose}>
    <section className="h-fit w-full max-w-xl overflow-hidden rounded-2xl border border-white/10 bg-[#181b18] shadow-2xl" onClick={(event)=>event.stopPropagation()}>
      <input autoFocus value={query} onInput={(event)=>setQuery(event.currentTarget.value)} onKeyDown={(event)=>{if(event.key==='Escape')onClose();}} placeholder="Search documents…" className="w-full border-b border-white/[.07] bg-transparent px-5 py-4 text-sm outline-none placeholder:text-zinc-600" />
      <div className="max-h-80 overflow-auto p-2">{matches.map((item)=><button key={item.id} onClick={()=>{onOpen(item.id);onClose();}} className="block w-full rounded-xl px-3 py-3 text-left hover:bg-white/5"><span className="block text-sm text-zinc-200">{item.title}</span><span className="mt-1 line-clamp-1 block text-xs text-zinc-600">{item.excerpt||'Empty document'}</span></button>)}{!matches.length?<p className="px-3 py-8 text-center text-xs text-zinc-600">No matching documents</p>:null}</div>
      <footer className="border-t border-white/[.07] px-4 py-2 text-[10px] text-zinc-600">Esc to close · ⌘K to open</footer>
    </section>
  </div>;
}
