export interface Document {
  id: string;
  title: string;
  content: string;
  excerpt?: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export type PatchStatus = 'proposed' | 'applied' | 'rejected';

export interface ContextItem {
  kind: 'ancestor' | 'reference' | 'backlink' | 'semantic';
  title: string;
  documentId: string;
  documentTitle: string;
  sectionId: string;
  content: string;
  score?: number;
}

export interface Patch {
  id: string;
  documentId: string;
  baseVersion: number;
  start: number;
  end: number;
  original: string;
  replacement: string;
  instruction: string;
  status: PatchStatus;
  createdAt: string;
  appliedAt?: string;
  context: ContextItem[];
  assessment: ContextAssessment;
  impacts: ImpactFinding[];
}

export interface ImpactFinding {
  kind: 'backlink' | 'semantic' | 'diagram';
  documentId: string;
  documentTitle: string;
  sectionId: string;
  title: string;
  reason: string;
  score?: number;
}

export interface ContextAssessment {
  score: number;
  level: 'low' | 'medium' | 'high';
  structural: number;
  explicit: number;
  semantic: number;
  unresolved: number;
  summary: string;
}
