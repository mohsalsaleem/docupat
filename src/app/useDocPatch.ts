import { useEffect, useRef, useState } from 'octane';
import type { EditorView } from '@codemirror/view';
import { documentApi, type Document, type Patch, type DocumentTemplate } from '../features/documents';
import { replaceEditorContent } from '../features/editor/codeMirror';

export function useDocPatch() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [document, setDocument] = useState<Document | null>(null);
  const [draft, setDraft] = useState('');
  const [selection, setSelection] = useState({ start: 0, end: 0 });
  const [instruction, setInstruction] = useState('');
  const [proposal, setProposal] = useState<Patch | null>(null);
  const [patches, setPatches] = useState<Patch[]>([]);
  const [busy, setBusy] = useState(false);
  const [saveState, setSaveState] = useState<'saved'|'saving'|'unsaved'|'error'>('saved');
  const [deleted, setDeleted] = useState<Document | null>(null);
  const [message, setMessage] = useState('Select text or a section, then describe what you want to change.');
  const [preview, setPreview] = useState(false);
  const editor = useRef<EditorView | null>(null);

  async function open(id: string) {
    try {
      const next = await documentApi.get(id);
      setDocument(next);
      setDraft(next.content);
      setProposal(null);
      setPatches(await documentApi.patches(id));
      setSaveState('saved');
      localStorage.setItem('stellarity:last-document', id);
      if (editor.current) replaceEditorContent(editor.current, next.content);
    } catch (error) {
      setMessage(asMessage(error));
    }
  }

  useEffect(() => {
    documentApi.list().then((items) => {
      setDocuments(items);
      const remembered = localStorage.getItem('stellarity:last-document');
      const initial = items.find((item) => item.id === remembered) ?? items[0];
      if (initial) open(initial.id);
    }).catch((error) => setMessage(asMessage(error)));
  }, []);

  async function save() {
    if (!document) return;
    setSaveState('saving');
    await run(async () => {
      const saved = await documentApi.save({ ...document, content: draft });
      setDocument(saved);
      setDocuments(await documentApi.list());
      setSaveState('saved');
      setMessage('All changes saved.');
    });
  }

  async function create(title:string, template:DocumentTemplate) {
    await run(async()=>{ const next=await documentApi.create(title,template.content(title)); setDocuments(await documentApi.list()); await open(next.id); setMessage(`Created ${template.name}.`); });
  }

  async function rename(title:string) {
    if (!document || !title.trim()) return;
    await run(async()=>{ const saved=await documentApi.save({...document,title:title.trim(),content:draft}); setDocument(saved); setDocuments(await documentApi.list()); setSaveState('saved'); setMessage('Document renamed.'); });
  }

  async function duplicate() {
    if (!document) return;
    await run(async()=>{ const next=await documentApi.create(`${document.title} copy`,draft); setDocuments(await documentApi.list()); await open(next.id); setMessage('Document duplicated.'); });
  }

  async function remove() {
    if (!document) return;
    await run(async()=>{ const removed=await documentApi.delete(document.id); setDeleted(removed); const items=await documentApi.list(); setDocuments(items); setDocument(null); setDraft(''); if(items[0]) await open(items[0].id); setMessage('Document deleted.'); });
  }

  async function undoDelete() {
    if (!deleted) return;
    await run(async()=>{ const restored=await documentApi.create(deleted.title,deleted.content); setDeleted(null); setDocuments(await documentApi.list()); await open(restored.id); setMessage('Document restored.'); });
  }

  async function restore(version:number) {
    if(!document)return;
    await run(async()=>{const restored=await documentApi.restore(document.id,version);setDocument(restored);setDraft(restored.content);setPatches(await documentApi.patches(restored.id));setDocuments(await documentApi.list());if(editor.current)replaceEditorContent(editor.current,restored.content);setMessage(`Restored version ${version}.`);});
  }

  useEffect(()=>{ if(!document||draft===document.content||busy)return; const timer=setTimeout(()=>save(),1200); return()=>clearTimeout(timer); },[draft,document?.id,document?.version,busy]);

  async function propose() {
    if (!document || !(selection.end > selection.start) || !instruction.trim()) return;
    await run(async () => {
      if (draft !== document.content) {
        setMessage('Let the current edit finish saving, then try again.');
        return;
      }
      setMessage('Writing a suggestion…');
      const next = await documentApi.propose(document, selection.start, selection.end, instruction, true);
      setProposal(next);
      setMessage('Your suggestion is ready to review.');
    });
  }

  async function apply() {
    if (!proposal) return;
    await run(async () => {
      const next = await documentApi.apply(proposal.id);
      setDocument(next);
      setDraft(next.content);
      if (editor.current) replaceEditorContent(editor.current, next.content);
      setProposal(null);
      setPatches(await documentApi.patches(next.id));
      setMessage('Change applied.');
    });
  }

  async function reject() {
    if (!proposal) return;
    await run(async () => {
      await documentApi.reject(proposal.id);
      setProposal(null);
      setPatches(await documentApi.patches(proposal.documentId));
      setMessage('Suggestion discarded. Your document was not changed.');
    });
  }

  async function regenerate() {
    if (!proposal || !document || !instruction.trim()) return;
    await run(async()=>{ await documentApi.reject(proposal.id); setMessage('Writing another suggestion…'); const next=await documentApi.propose(document,selection.start,selection.end,instruction,true); setProposal(next); setPatches(await documentApi.patches(document.id)); setMessage('A new suggestion is ready.'); });
  }

  function chooseSection(start: number, end: number) {
    editor.current?.dispatch({ selection: { anchor: start, head: end }, scrollIntoView: true });
    editor.current?.focus();
  }

  async function run(action: () => Promise<void>) {
    setBusy(true);
    try { await action(); } catch (error) { if(saveState === 'saving') setSaveState('error'); setMessage(asMessage(error)); } finally { setBusy(false); }
  }

  return {
    documents, document, draft, selection, instruction, proposal, patches, busy, message, preview, saveState, deleted,
    open, create, save, rename, duplicate, remove, undoDelete, restore, propose, apply, reject, regenerate, chooseSection,
    onEditor: (view: EditorView | null) => { editor.current = view; },
    onSelection: (start: number, end: number) => setSelection({ start, end }),
    onDraft: (value:string) => { setDraft(value); setSaveState(value === document?.content ? 'saved' : 'unsaved'); },
    setInstruction, setPreview,
  };
}

function asMessage(error: unknown) {
  const message = error instanceof Error ? error.message : '';
  if (/fetch|network|connect|model|llm/i.test(message)) return 'Could not reach the writing model. Check its connection and try again.';
  return message || 'Something went wrong. Please try again.';
}

export type DocPatchController = ReturnType<typeof useDocPatch>;
