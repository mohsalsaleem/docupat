export interface Heading { title: string; level: number; start: number; end: number }
export interface MermaidBlock { id: string; source: string }

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
