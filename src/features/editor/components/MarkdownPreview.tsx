import { useEffect, useRef } from 'octane';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { markdownParts } from '../markdown';
import { MermaidPreview } from './MermaidPreview';

export function MarkdownPreview({ content }: { content: string }) {
  const parts = markdownParts(content);
  return <section className="h-full min-w-0 overflow-auto">
    <article className="mx-auto max-w-4xl px-12 pb-40 pt-12">
      {parts.map((part) => part.type === 'mermaid'
        ? <div key={part.id} className="my-8"><MermaidPreview id={part.id} source={part.content} /></div>
        : <MarkdownSection key={part.id} source={part.content} />)}
    </article>
  </section>;
}

function MarkdownSection({ source }: { source: string }) {
  const host = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (!host.current) return;
    const rendered = marked.parse(source, { gfm: true, breaks: false });
    host.current.innerHTML = DOMPurify.sanitize(rendered as string);
  });
  return <div ref={host} className="markdown-preview" />;
}
