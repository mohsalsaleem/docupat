import { useEffect, useRef, useState } from 'octane';
import type { EditorView } from '@codemirror/view';
import { documentApi, type Document, type Patch } from '../features/documents';
import { replaceEditorContent } from '../features/editor/codeMirror';

export function useDocPatch() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [document, setDocument] = useState<Document | null>(null);
  const [draft, setDraft] = useState('');
  const [selection, setSelection] = useState({ start: 0, end: 0 });
  const [instruction, setInstruction] = useState('');
  const [useContext, setUseContext] = useState(true);
  const [proposal, setProposal] = useState<Patch | null>(null);
  const [patches, setPatches] = useState<Patch[]>([]);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('Select text or a section to create a scoped patch.');
  const [preview, setPreview] = useState(false);
  const editor = useRef<EditorView | null>(null);

  async function open(id: string) {
    try {
      const next = await documentApi.get(id);
      setDocument(next);
      setDraft(next.content);
      setProposal(null);
      setPatches(await documentApi.patches(id));
      if (editor.current) replaceEditorContent(editor.current, next.content);
    } catch (error) {
      setMessage(asMessage(error));
    }
  }

  useEffect(() => {
    documentApi.list().then((items) => {
      setDocuments(items);
      if (items[0]) open(items[0].id);
    }).catch((error) => setMessage(asMessage(error)));
  }, []);

  async function save() {
    if (!document) return;
    await run(async () => {
      const saved = await documentApi.save({ ...document, content: draft });
      setDocument(saved);
      setDocuments(await documentApi.list());
      setMessage(`Saved version ${saved.version}.`);
    });
  }

  async function propose() {
    if (!document || !(selection.end > selection.start) || !instruction.trim()) return;
    await run(async () => {
      if (draft !== document.content) {
        setMessage('Save the document before generating a patch.');
        return;
      }
      setMessage('The configured model is drafting a replacement…');
      const next = await documentApi.propose(document, selection.start, selection.end, instruction, useContext);
      setProposal(next);
      setMessage('Review the source diff and rendered diagrams before applying.');
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
      setMessage(`Patch applied as version ${next.version}.`);
    });
  }

  async function reject() {
    if (!proposal) return;
    await run(async () => {
      await documentApi.reject(proposal.id);
      setProposal(null);
      setPatches(await documentApi.patches(proposal.documentId));
      setMessage('Patch rejected; the document was not changed.');
    });
  }

  function chooseSection(start: number, end: number) {
    editor.current?.dispatch({ selection: { anchor: start, head: end }, scrollIntoView: true });
    editor.current?.focus();
  }

  async function run(action: () => Promise<void>) {
    setBusy(true);
    try { await action(); } catch (error) { setMessage(asMessage(error)); } finally { setBusy(false); }
  }

  return {
    documents, document, draft, selection, instruction, useContext, proposal, patches, busy, message, preview,
    open, save, propose, apply, reject, chooseSection,
    onEditor: (view: EditorView | null) => { editor.current = view; },
    onSelection: (start: number, end: number) => setSelection({ start, end }),
    onDraft: setDraft,
    setInstruction, setUseContext, setPreview,
  };
}

function asMessage(error: unknown) { return error instanceof Error ? error.message : 'Unexpected error'; }

export type DocPatchController = ReturnType<typeof useDocPatch>;
