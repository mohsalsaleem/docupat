import { Header } from './components/Header';
import { PatchPanel } from './components/PatchPanel';
import { Sidebar } from './components/Sidebar';
import { Workspace } from './components/Workspace';
import { useDocPatch } from './hooks/useDocPatch';

export function App() {
  const app = useDocPatch();
  return <div className="h-screen overflow-hidden bg-[#0b0d10] text-zinc-100">
    <Header preview={app.preview} dirty={Boolean(app.document && app.draft !== app.document.content)} busy={app.busy} onTogglePreview={() => app.setPreview(!app.preview)} onSave={app.save} />
    <main className="grid h-[calc(100vh-4rem)] grid-cols-[230px_minmax(420px,1fr)_390px]">
      <Sidebar documents={app.documents} active={app.document} draft={app.draft} patches={app.patches} onOpen={app.open} onSelect={app.chooseSection} />
      <Workspace preview={app.preview} content={app.proposal?.replacement ?? app.draft} editorHost={app.editorHost} />
      <PatchPanel selection={app.selection} instruction={app.instruction} useContext={app.useContext} proposal={app.proposal} busy={app.busy} message={app.message} onInstruction={app.setInstruction} onUseContext={app.setUseContext} onPropose={app.propose} onApply={app.apply} onReject={app.reject} />
    </main>
  </div>;
}
