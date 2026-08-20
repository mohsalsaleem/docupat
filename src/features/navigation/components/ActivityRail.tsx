import type { NavigationView } from '../types';

interface ActivityRailProps {
  active: NavigationView;
  onChange: (view: NavigationView) => void;
}

const items: Array<{ view: NavigationView; label: string; icon: string }> = [
  { view: 'files', label: 'Files', icon: 'files' },
  { view: 'outline', label: 'Outline', icon: 'outline' },
  { view: 'changes', label: 'Changes', icon: 'history' },
];

export function ActivityRail(props: ActivityRailProps) {
  return <nav className="flex min-h-0 flex-col items-center border-r border-white/[.07] bg-[var(--surface-1)] py-2" aria-label="Workspace views">
    {items.map((item) => <button key={item.view} title={item.label} aria-label={item.label} aria-pressed={props.active === item.view} onClick={() => props.onChange(item.view)} className={`relative grid size-10 place-items-center rounded-lg transition-colors ${props.active === item.view ? 'bg-white/[.07] text-[var(--accent)]' : 'text-zinc-500 hover:bg-white/[.04] hover:text-zinc-300'}`}>
      {props.active === item.view ? <span className="absolute -left-2 h-5 w-0.5 rounded-r bg-[var(--accent)]" /> : null}
      <RailIcon name={item.icon} />
    </button>)}
    <div className="mt-auto"><button title="Settings" aria-label="Settings" className="grid size-10 place-items-center rounded-lg text-zinc-500 transition-colors hover:bg-white/[.04] hover:text-zinc-300"><RailIcon name="settings" /></button></div>
  </nav>;
}

function RailIcon(props: { name: string }) {
  const paths: Record<string, unknown> = {
    files: <><path d="M4 5.5h5l1.5 2H20v11H4z" /><path d="M4 9h16" /></>,
    outline: <><path d="M5 6h14M5 12h10M5 18h7" /><circle cx="3" cy="6" r=".5" /><circle cx="3" cy="12" r=".5" /><circle cx="3" cy="18" r=".5" /></>,
    history: <><path d="M4 5v5h5" /><path d="M5.4 16.5A8 8 0 1 0 4 10" /><path d="M12 8v4l3 2" /></>,
    settings: <><circle cx="12" cy="12" r="3" /><path d="M19 13.5v-3l-2-.7-.7-1.7.9-1.9-2.2-2.1-1.8.9-1.8-.7L10.5 2h-3l-.7 2.3-1.7.7-1.9-.9-2.1 2.1.9 1.9-.7 1.7-2.3.7v3l2.3.7.7 1.7-.9 1.9 2.1 2.1 1.9-.9 1.7.7.7 2.3h3l.7-2.3 1.8-.7 1.8.9 2.2-2.1-.9-1.9.7-1.7z" /></>,
  };
  return <svg viewBox="0 0 24 24" className="size-[18px]" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">{paths[props.name]}</svg>;
}
