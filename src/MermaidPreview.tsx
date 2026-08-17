import { useEffect, useRef, useState } from 'octane';
import mermaid from 'mermaid';

mermaid.initialize({startOnLoad:false,securityLevel:'strict',theme:'dark'});
export function MermaidPreview({source,id}:{source:string;id:string}){
  const ref=useRef<HTMLDivElement|null>(null);const [error,setError]=useState('');
  useEffect(()=>{let active=true;(async()=>{try{await mermaid.parse(source);const {svg}=await mermaid.render(`docpatch-${id}-${Date.now()}`,source);if(active&&ref.current){ref.current.innerHTML=svg;setError('')}}catch(e){if(active)setError(e instanceof Error?e.message:'Invalid Mermaid diagram')}})();return()=>{active=false}},[source,id]);
  return <div className="rounded-xl border border-white/10 bg-[#111419] p-4">{error?<pre className="whitespace-pre-wrap text-xs text-red-300">{error}</pre>:<div ref={ref} className="flex min-h-48 items-center justify-center overflow-auto"/>}</div>;
}
