import { cn } from '../lib/cn';

export function SectionLabel(props: { children: unknown; spaced?: boolean }) {
  return <div className={cn('mb-3 text-[10px] font-bold uppercase tracking-[.18em] text-zinc-600', props.spaced && 'mt-7')}>{props.children}</div>;
}
