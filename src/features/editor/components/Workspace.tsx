import { useEffect, useEffectEvent, useRef } from 'octane';
import type { EditorView } from '@codemirror/view';
import { createEditor, replaceEditorContent } from '../codeMirror';
import { MarkdownPreview } from './MarkdownPreview';

interface WorkspaceProps { preview:boolean; content:string; onChange:(content:string)=>void; onSelection:(start:number,end:number)=>void; onEditor:(view:EditorView|null)=>void }

export function Workspace(props: WorkspaceProps) {
  if (!props.preview) return <EditorSurface content={props.content} onChange={props.onChange} onSelection={props.onSelection} onEditor={props.onEditor} />;
  return <MarkdownPreview content={props.content} />;
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
