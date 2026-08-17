export interface Document { id:string; title:string; content:string; version:number; createdAt:string; updatedAt:string }
export interface Patch { id:string; documentId:string; baseVersion:number; start:number; end:number; original:string; replacement:string; instruction:string; status:'proposed'|'applied'|'rejected'; createdAt:string; appliedAt?:string }

async function request<T>(path:string, init?:RequestInit):Promise<T>{
  const response=await fetch(path,{...init,headers:{'content-type':'application/json',...init?.headers}});
  const data=await response.json(); if(!response.ok) throw new Error(data.error||`Request failed: ${response.status}`); return data;
}
export const api={
  documents:()=>request<Document[]>('/api/documents'),
  document:(id:string)=>request<Document>(`/api/documents/${id}`),
  save:(doc:Document)=>request<Document>(`/api/documents/${doc.id}`,{method:'PUT',body:JSON.stringify({title:doc.title,content:doc.content,version:doc.version})}),
  patches:(id:string)=>request<Patch[]>(`/api/documents/${id}/patches`),
  propose:(doc:Document,start:number,end:number,instruction:string,useContext:boolean)=>request<Patch>(`/api/documents/${doc.id}/patches`,{method:'POST',body:JSON.stringify({start,end,version:doc.version,instruction,useContext})}),
  apply:(id:string)=>request<Document>(`/api/patches/${id}/apply`,{method:'POST'}),
  reject:(id:string)=>request<{status:string}>(`/api/patches/${id}/reject`,{method:'POST'}),
};
