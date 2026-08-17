import { useRef } from 'octane';
import type { EditorView } from '@codemirror/view';
import { mermaidBlocks } from '../diff';
import { createEditor } from '../editor';
import { MermaidPreview } from '../MermaidPreview';

interface WorkspaceProps { preview:boolean; content:string; onChange:(content:string)=>void; onSelection:(start:number,end:number)=>void; onEditor:(view:EditorView|null)=>void }

export function Workspace(props: WorkspaceProps) {
  if (!props.preview) return <EditorSurface content={props.content} onChange={props.onChange} onSelection={props.onSelection} onEditor={props.onEditor} />;
  const diagrams = mermaidBlocks(props.content);
  return <section className="min-w-0 overflow-hidden"><div className="h-full space-y-4 overflow-auto p-8">
    {diagrams.length ? diagrams.map((diagram) => <MermaidPreview key={diagram.id} {...diagram} />) : <div className="grid h-full place-items-center text-sm text-zinc-600">No Mermaid blocks in this document.</div>}
  </div></section>;
}

function EditorSurface(props:{content:string;onChange:(content:string)=>void;onSelection:(start:number,end:number)=>void;onEditor:(view:EditorView|null)=>void}) {
  const editor = useRef<EditorView | null>(null);
  function attach(host: HTMLDivElement | null) {
    if (host && !editor.current) {
      editor.current = createEditor(host, props.content, props.onChange, props.onSelection);
      props.onEditor(editor.current);
    } else if (!host && editor.current) {
      editor.current.destroy();
      editor.current = null;
      props.onEditor(null);
    }
  }
  return <section className="h-full min-w-0 overflow-hidden"><div ref={attach} className="h-full overflow-auto pb-32" /></section>;
}
