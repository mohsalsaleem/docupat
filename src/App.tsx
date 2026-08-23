import { useState } from 'octane';
import { useDocPatch } from './app/useDocPatch';
import { Workspace } from './features/editor';
import { ActivityRail, AppHeader, DocumentTabs, Sidebar, StatusBar, type NavigationView } from './features/navigation';
import { PatchComposer } from './features/patches';
import { NewDocumentDialog, type DocumentTemplate } from './features/documents';

export function App() {
  const app = useDocPatch();
  const [navigationView, setNavigationView] = useState<NavigationView>('files');
  const [creating,setCreating]=useState(false);
  const dirty = Boolean(app.document && app.draft !== app.document.content);
  const selected = Math.max(0, app.selection.end - app.selection.start);
  return <div className="flex h-screen flex-col overflow-hidden bg-[var(--surface-0)] text-zinc-100">
    <AppHeader preview={app.preview} dirty={dirty} busy={app.busy} onTogglePreview={() => app.setPreview(!app.preview)} onSave={app.save} />
    <main className="grid min-h-0 flex-1 grid-cols-[44px_250px_minmax(0,1fr)]">
      <ActivityRail active={navigationView} onChange={setNavigationView} />
      <Sidebar view={navigationView} documents={app.documents} active={app.document} draft={app.draft} patches={app.patches} onOpen={app.open} onSelect={app.chooseSection} onNew={()=>setCreating(true)} />
      <section className="flex min-w-0 flex-col overflow-hidden">
        <DocumentTabs document={app.document} dirty={dirty} />
        <div className="relative min-h-0 flex-1 overflow-hidden">
          <Workspace preview={app.preview} content={app.draft} onChange={app.onDraft} onSelection={app.onSelection} onEditor={app.onEditor} />
          <PatchComposer selection={app.selection} instruction={app.instruction} proposal={app.proposal} busy={app.busy} message={app.message} onInstruction={app.setInstruction} onPropose={app.propose} onApply={app.apply} onReject={app.reject} />
        </div>
      </section>
    </main>
    <StatusBar version={app.document?.version} selected={selected} busy={app.busy} message={app.message} />
    {creating?<NewDocumentDialog busy={app.busy} onClose={()=>setCreating(false)} onCreate={async(title:string,template:DocumentTemplate)=>{await app.create(title,template);setCreating(false);}}/>:null}
  </div>;
}
