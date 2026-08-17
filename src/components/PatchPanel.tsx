import type { Patch } from '../api';
import { DiffView } from './DiffView';

interface PatchPanelProps { selection:{start:number;end:number};instruction:string;useContext:boolean;proposal:Patch|null;busy:boolean;message:string;onInstruction:(value:string)=>void;onUseContext:(value:boolean)=>void;onPropose:()=>void;onApply:()=>void;onReject:()=>void }

export function PatchPanel(props: PatchPanelProps) {
  const selected = props.selection.end > props.selection.start;
  return <aside className="overflow-auto border-l border-white/10 bg-[#101318] p-5">
    <div className="text-[10px] font-bold uppercase tracking-[.18em] text-zinc-600">AI patch</div>
    <h2 className="mt-3 text-xl font-semibold">Change only what you select.</h2>
    <p className="mt-2 text-xs leading-5 text-zinc-500">Readable context and writable scope are separate.</p>
    <div className="mt-5 rounded-lg border border-white/10 bg-black/20 p-3 font-mono text-xs text-zinc-400">{selected ? `${props.selection.end - props.selection.start} characters selected` : 'No selection'}</div>
    <label className="mt-5 block text-xs font-semibold text-zinc-400">Instruction</label>
    <textarea value={props.instruction} onInput={(event) => props.onInstruction(event.currentTarget.value)} placeholder="Add passkey support and clarify recovery…" className="mt-2 h-28 w-full resize-none rounded-lg border border-white/10 bg-black/25 p-3 text-sm outline-none focus:border-lime-300/50" />
    <label className="mt-3 flex gap-2 text-xs leading-5 text-zinc-500"><input type="checkbox" checked={props.useContext} onChange={(event) => props.onUseContext(event.currentTarget.checked)} className="accent-lime-300" />Use the rest of the document as read-only context</label>
    <button disabled={props.busy || !selected || !props.instruction.trim()} onClick={props.onPropose} className="mt-4 w-full rounded-lg bg-lime-300 px-4 py-3 text-sm font-bold text-black disabled:opacity-30">{props.busy ? 'Working…' : 'Propose patch'}</button>
    <p className="mt-3 text-xs leading-5 text-zinc-500">{props.message}</p>
    {props.proposal ? <Proposal proposal={props.proposal} busy={props.busy} onApply={props.onApply} onReject={props.onReject} /> : null}
  </aside>;
}

function Proposal(props:{proposal:Patch;busy:boolean;onApply:()=>void;onReject:()=>void}) {
  return <section className="mt-6 border-t border-white/10 pt-5">
    <div className="mb-3 flex justify-between text-xs font-bold"><span>Proposed patch</span><span className="font-mono text-zinc-600">DIFF</span></div>
    <DiffView before={props.proposal.original} after={props.proposal.replacement} />
    <div className="mt-3 flex justify-end gap-2"><button onClick={props.onReject} className="rounded-lg border border-white/10 px-3 py-2 text-xs">Reject</button><button disabled={props.busy} onClick={props.onApply} className="rounded-lg bg-lime-300 px-3 py-2 text-xs font-bold text-black">Apply patch</button></div>
  </section>;
}
