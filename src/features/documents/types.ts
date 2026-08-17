export interface Document {
  id: string;
  title: string;
  content: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export type PatchStatus = 'proposed' | 'applied' | 'rejected';

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
}
