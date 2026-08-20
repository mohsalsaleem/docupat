export interface DocumentTemplate { id:string; name:string; description:string; content:(title:string)=>string }
export const documentTemplates: DocumentTemplate[] = [
  { id:'prd', name:'Product requirements', description:'Problem, goals, users, requirements, metrics, and open questions.', content:t=>`# ${t}\n\n## Problem\n\nDescribe the problem and why it matters.\n\n## Goals\n\n- \n\n## Users\n\n## Requirements\n\n### Functional\n\n### Non-functional\n\n## Success metrics\n\n## Open questions\n` },
  { id:'hld', name:'High-level design', description:'System context, components, data flow, reliability, and risks.', content:t=>`# ${t}\n\n## Context\n\n## Goals and constraints\n\n## Architecture\n\n\`\`\`mermaid\nflowchart LR\n    Client --> Service\n\`\`\`\n\n## Components\n\n## Data flow\n\n## Reliability and security\n\n## Risks and trade-offs\n` },
  { id:'adr', name:'Architecture decision', description:'Decision context, options, outcome, and consequences.', content:t=>`# ${t}\n\n**Status:** Proposed\n\n## Context\n\n## Decision drivers\n\n## Considered options\n\n## Decision\n\n## Consequences\n` },
  { id:'blank', name:'Blank Markdown', description:'Start with a title and an empty page.', content:t=>`# ${t}\n\n` },
];
