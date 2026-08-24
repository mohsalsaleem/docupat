import { useEffect, useEffectEvent, useRef } from 'octane';
import type { EditorView } from '@codemirror/view';
import { createEditor, replaceEditorContent } from '../codeMirror';
import { MarkdownPreview } from './MarkdownPreview';
import type { Document } from '../../documents';

interface WorkspaceProps { preview:boolean; documentId?:string; documents:Document[]; content:string; onChange:(content:string)=>void; onSelection:(start:number,end:number)=>void; onEditor:(view:EditorView|null)=>void; onOpenDocument:(id:string)=>void }

export function Workspace(props: WorkspaceProps) {
  if (!props.preview) return <EditorSurface documentId={props.documentId} content={props.content} onChange={props.onChange} onSelection={props.onSelection} onEditor={props.onEditor} />;
  return <MarkdownPreview documentId={props.documentId} documents={props.documents} onOpenDocument={props.onOpenDocument} content={props.content} />;
}

function EditorSurface(props:{documentId?:string;content:string;onChange:(content:string)=>void;onSelection:(start:number,end:number)=>void;onEditor:(view:EditorView|null)=>void}) {
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
    const scroller=view.scrollDOM;
    const key=`stellarity:editor-scroll:${props.documentId??'none'}`;
    scroller.scrollTop=Number(sessionStorage.getItem(key)??0);
    const remember=()=>sessionStorage.setItem(key,String(scroller.scrollTop));
    scroller.addEventListener('scroll',remember,{passive:true});
    return () => {
      scroller.removeEventListener('scroll',remember);
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
