import type { Patch } from '../../documents/types';
import { Button, Dialog } from '../../../ui';
import { DiffView } from './DiffView';
import { ContextSummary } from './ContextSummary';
import { ImpactSummary } from './ImpactSummary';

interface PatchPanelProps { selection:{start:number;end:number};instruction:string;proposal:Patch|null;busy:boolean;message:string;onInstruction:(value:string)=>void;onPropose:()=>void;onApply:()=>void;onReject:()=>void }

export function PatchComposer(props: PatchPanelProps) {
  const selected = props.selection.end > props.selection.start;
  return <>
    <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 flex justify-center bg-gradient-to-t from-[var(--surface-0)] to-transparent px-6 pb-5 pt-12">
      <div className="pointer-events-auto w-full max-w-3xl rounded-2xl border border-white/[.09] bg-[var(--surface-3)]/95 p-2 shadow-[0_18px_60px_rgba(0,0,0,.45)] backdrop-blur-xl">
        <div className="flex items-center gap-2">
          <span className={`ml-2 size-2 rounded-full ${selected ? 'bg-[var(--accent)] shadow-[0_0_10px_var(--accent-glow)]' : 'bg-zinc-600'}`} />
          <input value={props.instruction} onInput={(event) => props.onInstruction(event.currentTarget.value)} onKeyDown={(event) => { if (event.key === 'Enter' && !props.busy && selected && props.instruction.trim()) { event.preventDefault(); props.onPropose(); } }} placeholder={selected ? 'Describe the change to this selection…' : 'Select text or choose a section to edit with AI'} className="min-w-0 flex-1 bg-transparent px-2 py-3 text-sm outline-none placeholder:text-zinc-600" />
          <Button variant="primary" size="md" disabled={props.busy || !selected || !props.instruction.trim()} onClick={props.onPropose}>{props.busy ? 'Writing…' : 'Suggest edit'}</Button>
        </div>
        <div className="flex items-center justify-between px-3 pb-1 text-[11px] text-zinc-500">
          <span>{selected ? `${props.selection.end - props.selection.start} characters selected` : props.message}</span>
          <span>{selected ? 'Enter to suggest' : 'Your document stays unchanged until you approve'}</span>
        </div>
      </div>
    </div>
    {props.proposal ? <Proposal proposal={props.proposal} busy={props.busy} onApply={props.onApply} onReject={props.onReject} /> : null}
  </>;
}

function Proposal(props:{proposal:Patch;busy:boolean;onApply:()=>void;onReject:()=>void}) {
  return <Dialog title="Review suggested edit" description="Only the selected text will change" badge="AI SUGGESTION" busy={props.busy} confirmLabel="Apply change" dismissLabel="Discard" onConfirm={props.onApply} onDismiss={props.onReject}>
    <ContextSummary items={props.proposal.context ?? []} assessment={props.proposal.assessment} />
    <ImpactSummary items={props.proposal.impacts ?? []} />
    <DiffView before={props.proposal.original} after={props.proposal.replacement} />
  </Dialog>;
}
