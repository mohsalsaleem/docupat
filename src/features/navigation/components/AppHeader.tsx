import { Button } from '../../../ui';
import type { Document } from '../../documents';
import { DocumentActions } from './DocumentActions';

interface HeaderProps { preview:boolean; dirty:boolean; busy:boolean; saveState:'saved'|'saving'|'unsaved'|'error'; document:Document|null; content:string; onTogglePreview:()=>void; onSave:()=>void; onSearch:()=>void; onRename:(title:string)=>void; onDuplicate:()=>void; onDelete:()=>void }

export function AppHeader(props: HeaderProps) {
  return <header className="flex h-12 items-center justify-between border-b border-white/[.07] bg-[var(--surface-1)] px-3">
    <div className="flex items-center gap-3">
      <div className="grid size-7 place-items-center rounded-lg bg-[var(--accent)] text-xs font-black text-[#11130f] shadow-[0_0_18px_var(--accent-glow)]">S</div>
      <b className="text-sm font-semibold tracking-tight">Stellarity</b>
    </div>
    <div className="flex gap-2">
      <Button onClick={props.onSearch}>Search <span className="ml-2 text-zinc-600">⌘K</span></Button>
      <Button onClick={props.onTogglePreview}>{props.preview ? 'Edit' : 'Preview'}</Button>
      <Button variant="secondary" disabled={props.busy || !props.dirty} onClick={props.onSave}>{saveLabel(props.saveState)}</Button>
      <DocumentActions document={props.document} content={props.content} onRename={props.onRename} onDuplicate={props.onDuplicate} onDelete={props.onDelete} />
    </div>
  </header>;
}

function saveLabel(state:HeaderProps['saveState']) { return state==='saving'?'Saving…':state==='unsaved'?'Save':state==='error'?'Retry save':'Saved'; }
