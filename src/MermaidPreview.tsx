import { useEffect, useRef } from 'octane';
import mermaid from 'mermaid';

mermaid.initialize({startOnLoad:false,securityLevel:'strict',theme:'dark'});
export function MermaidPreview({source,id}:{source:string;id:string}){
  const ref = useRef<HTMLDivElement | null>(null);

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

  return <div className="rounded-xl border border-white/10 bg-[#111419] p-4"><div ref={ref} className="flex min-h-48 items-center justify-center overflow-auto" /></div>;
}
