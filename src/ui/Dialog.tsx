import { Button } from './Button';

interface DialogProps {
  title: string;
  description?: string;
  badge?: string;
  children: unknown;
  busy?: boolean;
  confirmLabel: string;
  dismissLabel: string;
  onConfirm: () => void;
  onDismiss: () => void;
}

export function Dialog(props: DialogProps) {
  return <div className="fixed inset-0 z-30 grid place-items-center bg-black/45 p-8 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={props.title}>
    <section className="max-h-[75vh] w-full max-w-3xl overflow-auto rounded-2xl border border-white/10 bg-[#171a20] p-5 shadow-2xl">
      <header className="mb-4 flex items-center justify-between">
        <div><div className="text-sm font-semibold">{props.title}</div>{props.description ? <div className="mt-1 text-xs text-zinc-500">{props.description}</div> : null}</div>
        {props.badge ? <span className="rounded-full bg-lime-300/10 px-2 py-1 font-mono text-[10px] text-lime-200">{props.badge}</span> : null}
      </header>
      {props.children}
      <footer className="mt-5 flex justify-end gap-2">
        <Button onClick={props.onDismiss}>{props.dismissLabel}</Button>
        <Button variant="primary" size="md" disabled={props.busy} onClick={props.onConfirm}>{props.confirmLabel}</Button>
      </footer>
    </section>
  </div>;
}
