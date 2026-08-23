export interface Heading { title: string; level: number; start: number; end: number }
export interface MermaidBlock { id: string; source: string }
export interface MarkdownPart { id: string; type: 'markdown' | 'mermaid'; content: string }

export function headings(markdown: string): Heading[] {
  const matches = [...markdown.matchAll(/^(#{1,6})\s+(.+)$/gm)].map((match) => ({
    title: match[2].trim(), level: match[1].length, start: match.index!,
  }));
  return matches.map((heading, index) => ({
    ...heading,
    end: matches.slice(index + 1).find((next) => next.level <= heading.level)?.start ?? markdown.length,
  }));
}

export function mermaidBlocks(markdown: string): MermaidBlock[] {
  return [...markdown.matchAll(/```mermaid\s*\n([\s\S]*?)```/g)].map((match, index) => ({
    id: `diagram-${index}`,
    source: match[1].trim(),
  }));
}

export function markdownParts(markdown: string): MarkdownPart[] {
  const parts: MarkdownPart[] = [];
  const pattern = /```mermaid\s*\n([\s\S]*?)```/g;
  let cursor = 0;
  let index = 0;
  for (const match of markdown.matchAll(pattern)) {
    const start = match.index!;
    if (start > cursor) parts.push({ id: `markdown-${index}`, type: 'markdown', content: markdown.slice(cursor, start) });
    parts.push({ id: `mermaid-${index}`, type: 'mermaid', content: match[1].trim() });
    cursor = start + match[0].length;
    index++;
  }
  if (cursor < markdown.length || !parts.length) parts.push({ id: `markdown-${index}`, type: 'markdown', content: markdown.slice(cursor) });
  return parts;
}
