import { useEffect, useRef, useState } from 'octane';
import mermaid from 'mermaid';

mermaid.initialize({startOnLoad:false,securityLevel:'strict',theme:'dark'});
export function MermaidPreview({source,id}:{source:string;id:string}){
  const ref = useRef<HTMLDivElement | null>(null);
  const [copied,setCopied]=useState(false);

  async function copySource(){await navigator.clipboard.writeText(source);setCopied(true);setTimeout(()=>setCopied(false),1200);}
  function exportSVG(){const svg=ref.current?.querySelector('svg')?.outerHTML;if(!svg)return;const link=document.createElement('a');link.href=URL.createObjectURL(new Blob([svg],{type:'image/svg+xml'}));link.download=`${id}.svg`;link.click();URL.revokeObjectURL(link.href);}

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        await mermaid.parse(source);
        const { svg } = await mermaid.render(`docpatch-${id}-${Date.now()}`, source);
        if (active && ref.current) {
          ref.current.innerHTML = svg;
          ref.current.classList.remove('whitespace-pre-wrap', 'text-xs', 'text-red-300');
          ref.current.classList.add('flex', 'items-center', 'justify-center', 'overflow-auto');
        }
      } catch (cause) {
        if (active && ref.current) {
          ref.current.textContent = cause instanceof Error ? cause.message : 'Invalid Mermaid diagram';
          ref.current.classList.remove('flex', 'items-center', 'justify-center', 'overflow-auto');
          ref.current.classList.add('whitespace-pre-wrap', 'text-xs', 'text-red-300');
        }
      }
    })();
    return () => { active = false; };
  });

  return <div className="rounded-xl border border-white/10 bg-[#111419] p-4"><div className="mb-2 flex justify-end gap-2"><button onClick={copySource} className="rounded-md px-2 py-1 text-[10px] text-zinc-500 hover:bg-white/5 hover:text-zinc-300">{copied?'Copied':'Copy source'}</button><button onClick={exportSVG} className="rounded-md px-2 py-1 text-[10px] text-zinc-500 hover:bg-white/5 hover:text-zinc-300">Export SVG</button></div><div ref={ref} className="flex min-h-48 items-center justify-center overflow-auto" /></div>;
}
