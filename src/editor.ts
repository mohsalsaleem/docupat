import { EditorState, type Extension } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine, drawSelection } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { markdown } from '@codemirror/lang-markdown';
import { oneDark } from '@codemirror/theme-one-dark';

export function createEditor(parent:HTMLElement,content:string,onChange:(value:string)=>void,onSelection:(start:number,end:number)=>void){
  const listener=EditorView.updateListener.of(update=>{if(update.docChanged)onChange(update.state.doc.toString());if(update.selectionSet||update.docChanged){const r=update.state.selection.main;onSelection(r.from,r.to)}});
  const extensions:Extension[]=[lineNumbers(),highlightActiveLine(),drawSelection(),history(),markdown(),oneDark,keymap.of([...defaultKeymap,...historyKeymap]),listener,EditorView.lineWrapping];
  return new EditorView({parent,state:EditorState.create({doc:content,extensions})});
}
export function replaceEditorContent(view:EditorView,content:string){view.dispatch({changes:{from:0,to:view.state.doc.length,insert:content}})}
