# Architecture

DocPatch uses ports and adapters so domain behavior does not depend on delivery or infrastructure choices.

## Backend

```text
main (composition root)
  ├── httpapi.Handler      HTTP transport
  ├── document.Service     use cases and orchestration
  │     ├── Repository     persistence port
  │     └── Generator      language-model port
  ├── storage.Repository   SQLite adapter
  ├── llm.Client           provider-neutral model adapter
  └── domain               entities and scoped-patch invariants
```

- `internal/domain` owns document and patch entities, UTF-16 offset conversion, and the only operation allowed to replace document content.
- `internal/document` owns application use cases. It depends on interfaces it defines, not SQLite or HTTP.
- `internal/storage` implements the repository port with transactions and optimistic versions.
- `internal/llm` implements the generator port, owns prompt construction, and selects an OpenAI-compatible or Anthropic adapter from configuration.
- `internal/httpapi` translates JSON/HTTP requests into application inputs and maps domain errors to status codes.
- `main.go` is only the composition root, environment configuration, embedded static assets, and sample seeding.

## Frontend

```text
App
  ├── useDocPatch controller
  │     ├── API client
  │     └── CodeMirror adapter
  ├── Header
  ├── Sidebar
  ├── Workspace
  │     └── MermaidPreview
  └── PatchPanel
        └── DiffView
```

- Components render one area and communicate through typed props.
- `useDocPatch` owns application state and async workflows.
- `api.ts` owns transport details.
- `editor.ts` isolates CodeMirror integration.
- `diff.ts` contains pure parsing and diff functions.
- `MermaidPreview` isolates Mermaid lifecycle and validation.

## Core invariant

The model returns replacement text but never controls the replacement location. Applying a patch loads the authoritative SQLite document, verifies its version and original selected text, converts CodeMirror UTF-16 positions into UTF-8 byte boundaries, and replaces only that range inside one transaction.
