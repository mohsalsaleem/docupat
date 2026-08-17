import type { Patch } from '../api';
import { DiffView } from './DiffView';

interface PatchPanelProps { selection:{start:number;end:number};instruction:string;useContext:boolean;proposal:Patch|null;busy:boolean;message:string;onInstruction:(value:string)=>void;onUseContext:(value:boolean)=>void;onPropose:()=>void;onApply:()=>void;onReject:()=>void }

export function PatchPanel(props: PatchPanelProps) {
  const selected = props.selection.end > props.selection.start;
  return <>
    <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 flex justify-center bg-gradient-to-t from-[#0b0d10] via-[#0b0d10]/90 to-transparent px-6 pb-5 pt-12">
      <div className="pointer-events-auto w-full max-w-3xl rounded-2xl border border-white/10 bg-[#171a20]/95 p-2 shadow-2xl backdrop-blur-xl">
        <div className="flex items-center gap-2">
          <span className={`ml-2 size-2 rounded-full ${selected ? 'bg-lime-300' : 'bg-zinc-600'}`} />
          <input value={props.instruction} onInput={(event) => props.onInstruction(event.currentTarget.value)} onKeyDown={(event) => { if (event.key === 'Enter' && !props.busy && selected && props.instruction.trim()) { event.preventDefault(); props.onPropose(); } }} placeholder={selected ? 'Describe the change to this selection…' : 'Select text or choose a section to edit with AI'} className="min-w-0 flex-1 bg-transparent px-2 py-3 text-sm outline-none placeholder:text-zinc-600" />
          <button disabled={props.busy || !selected || !props.instruction.trim()} onClick={props.onPropose} className="rounded-xl bg-lime-300 px-4 py-2.5 text-xs font-bold text-black disabled:bg-white/5 disabled:text-zinc-600">{props.busy ? 'Drafting…' : 'Generate'}</button>
        </div>
        <div className="flex items-center justify-between px-3 pb-1 text-[11px] text-zinc-500">
          <span>{selected ? `${props.selection.end - props.selection.start} characters selected` : props.message}</span>
          <label className="flex items-center gap-2"><input type="checkbox" checked={props.useContext} onChange={(event) => props.onUseContext(event.currentTarget.checked)} className="accent-lime-300" />Read document context</label>
        </div>
      </div>
    </div>
    {props.proposal ? <Proposal proposal={props.proposal} busy={props.busy} onApply={props.onApply} onReject={props.onReject} /> : null}
  </>;
}

function Proposal(props:{proposal:Patch;busy:boolean;onApply:()=>void;onReject:()=>void}) {
  return <div className="absolute inset-0 z-30 grid place-items-center bg-black/45 p-8 backdrop-blur-sm">
    <section className="max-h-[75vh] w-full max-w-3xl overflow-auto rounded-2xl border border-white/10 bg-[#171a20] p-5 shadow-2xl">
      <div className="mb-4 flex items-center justify-between"><div><div className="text-sm font-semibold">Review focused edit</div><div className="mt-1 text-xs text-zinc-500">Only the selected text will change</div></div><span className="rounded-full bg-lime-300/10 px-2 py-1 font-mono text-[10px] text-lime-200">SCOPED PATCH</span></div>
      <DiffView before={props.proposal.original} after={props.proposal.replacement} />
      <div className="mt-5 flex justify-end gap-2"><button onClick={props.onReject} className="rounded-lg border border-white/10 px-4 py-2.5 text-xs hover:bg-white/5">Discard</button><button disabled={props.busy} onClick={props.onApply} className="rounded-lg bg-lime-300 px-4 py-2.5 text-xs font-bold text-black">Apply change</button></div>
    </section>
  </div>;
}
