import { useEffect, useState } from 'octane';
import { useDocPatch } from './app/useDocPatch';
import { Workspace } from './features/editor';
import { ActivityRail, AppHeader, DocumentTabs, SearchPalette, Sidebar, StatusBar, UndoToast, type NavigationView } from './features/navigation';
import { PatchComposer } from './features/patches';
import { NewDocumentDialog, type DocumentTemplate } from './features/documents';

export function App() {
  const app = useDocPatch();
  const [navigationView, setNavigationView] = useState<NavigationView>('files');
  const [creating,setCreating]=useState(false);
  const [searching,setSearching]=useState(false);
  const dirty = Boolean(app.document && app.draft !== app.document.content);
  const selected = Math.max(0, app.selection.end - app.selection.start);
  const words = app.draft.trim() ? app.draft.trim().split(/\s+/).length : 0;
  useEffect(()=>{
    const onKey=(event:KeyboardEvent)=>{
      const command=event.metaKey||event.ctrlKey;
      if(!command)return;
      if(event.key.toLowerCase()==='s'){event.preventDefault();app.save();}
      if(event.key.toLowerCase()==='k'){event.preventDefault();setSearching(true);}
      if(event.key.toLowerCase()==='n'){event.preventDefault();setCreating(true);}
      if(event.shiftKey&&event.key.toLowerCase()==='p'){event.preventDefault();app.setPreview(!app.preview);}
    };
    window.addEventListener('keydown',onKey);
    return()=>window.removeEventListener('keydown',onKey);
  });
  return <div className="flex h-screen flex-col overflow-hidden bg-[var(--surface-0)] text-zinc-100">
    <AppHeader preview={app.preview} dirty={dirty} busy={app.busy} saveState={app.saveState} document={app.document} content={app.draft} onTogglePreview={() => app.setPreview(!app.preview)} onSave={app.save} onSearch={()=>setSearching(true)} onRename={app.rename} onDuplicate={app.duplicate} onDelete={app.remove} />
    <main className="grid min-h-0 flex-1 grid-cols-[44px_250px_minmax(0,1fr)]">
      <ActivityRail active={navigationView} onChange={setNavigationView} />
      <Sidebar view={navigationView} documents={app.documents} active={app.document} draft={app.draft} patches={app.patches} onOpen={app.open} onSelect={app.chooseSection} onNew={()=>setCreating(true)} onRestore={app.restore} />
      <section className="flex min-w-0 flex-col overflow-hidden">
        <DocumentTabs document={app.document} dirty={dirty} />
        <div className="relative min-h-0 flex-1 overflow-hidden">
          <Workspace preview={app.preview} documentId={app.document?.id} documents={app.documents} content={app.draft} onChange={app.onDraft} onSelection={app.onSelection} onEditor={app.onEditor} onOpenDocument={app.open} />
          {!app.preview ? <PatchComposer selection={app.selection} instruction={app.instruction} proposal={app.proposal} busy={app.busy} message={app.message} onInstruction={app.setInstruction} onPropose={app.propose} onApply={app.apply} onReject={app.reject} onRegenerate={app.regenerate} /> : null}
        </div>
      </section>
    </main>
    <StatusBar version={app.document?.version} selected={selected} busy={app.busy} message={app.message} words={words} />
    {creating?<NewDocumentDialog busy={app.busy} onClose={()=>setCreating(false)} onCreate={async(title:string,template:DocumentTemplate)=>{await app.create(title,template);setCreating(false);}}/>:null}
    {searching?<SearchPalette documents={app.documents} onOpen={app.open} onClose={()=>setSearching(false)} />:null}
    {app.deleted?<UndoToast label={`Deleted “${app.deleted.title}”`} onUndo={app.undoDelete} />:null}
  </div>;
}
