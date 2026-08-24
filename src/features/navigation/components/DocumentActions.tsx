import { useState } from 'octane';
import type { Document } from '../../documents';
import { Button, Dialog } from '../../../ui';

interface Props { document:Document|null; content:string; onRename:(title:string)=>void; onDuplicate:()=>void; onDelete:()=>void }

export function DocumentActions(props:Props) {
  const [open,setOpen]=useState(false);
  const [renaming,setRenaming]=useState(false);
  const [title,setTitle]=useState('');
  const [copied,setCopied]=useState(false);
  async function copy() {
    if(!props.document)return;
    await navigator.clipboard.writeText(props.content);
    setCopied(true);
    setTimeout(()=>setCopied(false),1400);
  }
  function download() {
    if(!props.document)return;
    const blob=new Blob([props.content],{type:'text/markdown;charset=utf-8'});
    const link=document.createElement('a');
    link.href=URL.createObjectURL(blob);
    link.download=`${slug(props.document.title)}.md`;
    link.click();
    URL.revokeObjectURL(link.href);
  }
  return <div className="relative flex gap-2">
    <Button disabled={!props.document} onClick={copy}>{copied?'Copied':'Copy Markdown'}</Button>
    <Button disabled={!props.document} onClick={()=>setOpen(!open)} className="px-3" >•••</Button>
    {open?<div className="absolute right-0 top-11 z-30 w-48 rounded-xl border border-white/10 bg-[#1b1e1b] p-1.5 text-xs shadow-2xl">
      <MenuItem label="Download .md" onClick={()=>{download();setOpen(false);}} />
      <MenuItem label="Rename" onClick={()=>{setTitle(props.document?.title??'');setRenaming(true);setOpen(false);}} />
      <MenuItem label="Duplicate" onClick={()=>{props.onDuplicate();setOpen(false);}} />
      <div className="my-1 border-t border-white/[.07]" />
      <MenuItem label="Delete" danger onClick={()=>{props.onDelete();setOpen(false);}} />
    </div>:null}
    {renaming?<Dialog title="Rename document" confirmLabel="Rename" dismissLabel="Cancel" onConfirm={()=>{props.onRename(title);setRenaming(false);}} onDismiss={()=>setRenaming(false)}>
      <label className="block text-xs text-zinc-400">Document title<input value={title} onInput={(event)=>setTitle(event.currentTarget.value)} className="mt-2 w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2.5 text-sm text-white outline-none focus:border-lime-300/30" /></label>
    </Dialog>:null}
  </div>;
}

function MenuItem({label,onClick,danger=false}:{label:string;onClick:()=>void;danger?:boolean}) { return <button onClick={onClick} className={`block w-full rounded-lg px-3 py-2 text-left hover:bg-white/5 ${danger?'text-red-300':'text-zinc-300'}`}>{label}</button>; }
function slug(value:string){return value.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'')||'document';}
