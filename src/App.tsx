import { useDocPatch } from './app/useDocPatch';
import { Workspace } from './features/editor';
import { AppHeader, Sidebar } from './features/navigation';
import { PatchComposer } from './features/patches';

export function App() {
  const app = useDocPatch();
  return <div className="h-screen overflow-hidden bg-[#0b0d10] text-zinc-100">
    <AppHeader preview={app.preview} dirty={Boolean(app.document && app.draft !== app.document.content)} busy={app.busy} onTogglePreview={() => app.setPreview(!app.preview)} onSave={app.save} />
    <main className="grid h-[calc(100vh-4rem)] grid-cols-[230px_minmax(0,1fr)]">
      <Sidebar documents={app.documents} active={app.document} draft={app.draft} patches={app.patches} onOpen={app.open} onSelect={app.chooseSection} />
      <div className="relative min-w-0 overflow-hidden">
        <Workspace preview={app.preview} content={app.draft} onChange={app.onDraft} onSelection={app.onSelection} onEditor={app.onEditor} />
        <PatchComposer selection={app.selection} instruction={app.instruction} useContext={app.useContext} proposal={app.proposal} busy={app.busy} message={app.message} onInstruction={app.setInstruction} onUseContext={app.setUseContext} onPropose={app.propose} onApply={app.apply} onReject={app.reject} />
      </div>
    </main>
  </div>;
}
