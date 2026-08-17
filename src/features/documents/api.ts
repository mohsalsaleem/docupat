import { requestJSON } from '../../lib/http';
import type { Document, Patch } from './types';

export const documentApi = {
  list: () => requestJSON<Document[]>('/api/documents'),
  get: (id: string) => requestJSON<Document>(`/api/documents/${id}`),
  save: (document: Document) => requestJSON<Document>(`/api/documents/${document.id}`, {
    method: 'PUT',
    body: JSON.stringify({ title: document.title, content: document.content, version: document.version }),
  }),
  patches: (id: string) => requestJSON<Patch[]>(`/api/documents/${id}/patches`),
  propose: (document: Document, start: number, end: number, instruction: string, useContext: boolean) => requestJSON<Patch>(`/api/documents/${document.id}/patches`, {
    method: 'POST',
    body: JSON.stringify({ start, end, version: document.version, instruction, useContext }),
  }),
  apply: (id: string) => requestJSON<Document>(`/api/patches/${id}/apply`, { method: 'POST' }),
  reject: (id: string) => requestJSON<{ status: string }>(`/api/patches/${id}/reject`, { method: 'POST' }),
};
