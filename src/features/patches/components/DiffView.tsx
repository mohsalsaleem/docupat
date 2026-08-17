import { lineDiff } from '../diff';

export function DiffView(props:{before:string;after:string}) {
  return <div className="max-h-72 overflow-auto rounded-lg border border-white/10 bg-black/30 py-2 font-mono text-[11px]">
    {lineDiff(props.before, props.after).map((line, index) => <div key={index} className={`grid grid-cols-[20px_1fr] px-2 py-0.5 ${line.kind === 'add' ? 'bg-emerald-500/15 text-emerald-200' : line.kind === 'remove' ? 'bg-red-500/15 text-red-200' : 'text-zinc-600'}`}>
      <span>{line.kind === 'add' ? '+' : line.kind === 'remove' ? '−' : ' '}</span><span className="whitespace-pre-wrap">{line.text || ' '}</span>
    </div>)}
  </div>;
}
