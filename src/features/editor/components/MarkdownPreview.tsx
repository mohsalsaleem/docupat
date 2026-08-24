import { useEffect, useRef } from 'octane';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { markdownParts } from '../markdown';
import { MermaidPreview } from './MermaidPreview';
import type { Document } from '../../documents';

export function MarkdownPreview({ content, documentId, documents, onOpenDocument }: { content: string; documentId?:string; documents:Document[]; onOpenDocument:(id:string)=>void }) {
  const parts = markdownParts(content);
  const scrollHost=useRef<HTMLElement|null>(null);
  useEffect(()=>{
    const host=scrollHost.current;if(!host)return;
    const key=`stellarity:preview-scroll:${documentId??'none'}`;
    host.scrollTop=Number(sessionStorage.getItem(key)??0);
    const remember=()=>sessionStorage.setItem(key,String(host.scrollTop));
    host.addEventListener('scroll',remember,{passive:true});
    return()=>host.removeEventListener('scroll',remember);
  },[documentId]);
  return <section ref={scrollHost} className="h-full min-w-0 overflow-auto">
    <article className="mx-auto max-w-4xl px-12 pb-40 pt-12">
      {parts.map((part) => part.type === 'mermaid'
        ? <div key={part.id} className="my-8"><MermaidPreview id={part.id} source={part.content} /></div>
        : <MarkdownSection key={part.id} source={part.content} documents={documents} onOpenDocument={onOpenDocument} />)}
    </article>
  </section>;
}

function MarkdownSection({ source, documents, onOpenDocument }: { source: string; documents:Document[]; onOpenDocument:(id:string)=>void }) {
  const host = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (!host.current) return;
    const rendered = marked.parse(source, { gfm: true, breaks: false });
    host.current.innerHTML = DOMPurify.sanitize(rendered as string);
    host.current.querySelectorAll('h1,h2,h3,h4,h5,h6').forEach((heading)=>{heading.id=slug(heading.textContent??'section');});
    host.current.querySelectorAll('pre').forEach((block)=>{
      const code=block.querySelector('code')?.textContent??block.textContent??'';
      const button=document.createElement('button');button.className='copy-code';button.textContent='Copy';button.type='button';
      button.addEventListener('click',async()=>{await navigator.clipboard.writeText(code);button.textContent='Copied';setTimeout(()=>button.textContent='Copy',1200);});
      block.appendChild(button);
    });
    const navigate=(event:MouseEvent)=>{
      const anchor=(event.target as HTMLElement).closest('a');const href=anchor?.getAttribute('href');if(!href)return;
      if(href.startsWith('#')){event.preventDefault();document.getElementById(href.slice(1))?.scrollIntoView({behavior:'smooth'});return;}
      const [path,hash]=href.split('#');if(!/\.md$/i.test(path))return;
      const wanted=slug(decodeURIComponent(path).split('/').pop()!.replace(/\.md$/i,''));
      const match=documents.find((item)=>slug(item.title)===wanted);if(!match)return;
      event.preventDefault();onOpenDocument(match.id);if(hash)setTimeout(()=>document.getElementById(slug(decodeURIComponent(hash)))?.scrollIntoView({behavior:'smooth'}),100);
    };
    host.current.addEventListener('click',navigate);
    const current=host.current;
    return()=>current.removeEventListener('click',navigate);
  });
  return <div ref={host} className="markdown-preview" />;
}

function slug(value:string){return value.toLowerCase().trim().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'');}
