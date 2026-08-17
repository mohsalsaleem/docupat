import { mermaidBlocks } from '../diff';
import { MermaidPreview } from '../MermaidPreview';

interface WorkspaceProps { preview:boolean; content:string; editorHost:{current:HTMLDivElement|null} }

export function Workspace(props: WorkspaceProps) {
  if (!props.preview) return <section className="min-w-0 overflow-hidden"><div ref={props.editorHost} className="h-full overflow-auto" /></section>;
  const diagrams = mermaidBlocks(props.content);
  return <section className="min-w-0 overflow-hidden"><div className="h-full space-y-4 overflow-auto p-8">
    {diagrams.length ? diagrams.map((diagram) => <MermaidPreview key={diagram.id} {...diagram} />) : <div className="grid h-full place-items-center text-sm text-zinc-600">No Mermaid blocks in this document.</div>}
  </div></section>;
}
