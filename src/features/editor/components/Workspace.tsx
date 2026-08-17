import { useEffect, useEffectEvent, useRef } from 'octane';
import type { EditorView } from '@codemirror/view';
import { createEditor, replaceEditorContent } from '../codeMirror';
import { mermaidBlocks } from '../markdown';
import { MermaidPreview } from './MermaidPreview';

interface WorkspaceProps { preview:boolean; content:string; onChange:(content:string)=>void; onSelection:(start:number,end:number)=>void; onEditor:(view:EditorView|null)=>void }

export function Workspace(props: WorkspaceProps) {
  if (!props.preview) return <EditorSurface content={props.content} onChange={props.onChange} onSelection={props.onSelection} onEditor={props.onEditor} />;
  const diagrams = mermaidBlocks(props.content);
  return <section className="min-w-0 overflow-hidden"><div className="h-full space-y-4 overflow-auto p-8">
    {diagrams.length ? diagrams.map((diagram) => <MermaidPreview key={diagram.id} {...diagram} />) : <div className="grid h-full place-items-center text-sm text-zinc-600">No Mermaid blocks in this document.</div>}
  </div></section>;
}

function EditorSurface(props:{content:string;onChange:(content:string)=>void;onSelection:(start:number,end:number)=>void;onEditor:(view:EditorView|null)=>void}) {
  const host = useRef<HTMLDivElement | null>(null);
  const editor = useRef<EditorView | null>(null);
  const onChange = useEffectEvent(props.onChange);
  const onSelection = useEffectEvent(props.onSelection);
  const onEditor = useEffectEvent(props.onEditor);

  useEffect(() => {
    if (!host.current) return;
    const view = createEditor(host.current, props.content, onChange, onSelection);
    editor.current = view;
    onEditor(view);
    return () => {
      onEditor(null);
      editor.current = null;
      view.destroy();
    };
  }, []);

  useEffect(() => {
    const view = editor.current;
    if (view && view.state.doc.toString() !== props.content) replaceEditorContent(view, props.content);
  });

  return <section className="h-full min-w-0 overflow-hidden"><div ref={host} className="h-full overflow-auto pb-32" /></section>;
}
