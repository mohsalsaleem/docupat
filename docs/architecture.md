# Architecture

Stellarity uses ports and adapters so domain behavior does not depend on delivery or infrastructure choices.

## Backend

```text
main (composition root)
  ├── httpapi.Handler      HTTP transport
  ├── document.Service     use cases and orchestration
  │     ├── Repository     persistence port
  │     ├── Generator      language-model port
  │     └── ContextCompiler deterministic context port
  ├── contextcompile       Markdown relationship adapter
  ├── storage.Repository   SQLite adapter
  ├── llm.Client           provider-neutral model adapter
  └── domain               entities and scoped-patch invariants
```

- `internal/domain` owns document and patch entities, UTF-16 offset conversion, and the only operation allowed to replace document content.
- `internal/document` owns application use cases. It depends on interfaces it defines, not SQLite or HTTP.
- `internal/contextcompile` parses headings and explicit links, resolves ancestors/references/backlinks, deduplicates sources, and enforces a character budget without calling a model.
- `internal/storage` implements the repository port with transactions and optimistic versions.
- `internal/llm` implements the generator port, owns prompt construction, and selects an OpenAI-compatible or Anthropic adapter from configuration.
- `internal/httpapi` translates JSON/HTTP requests into application inputs and maps domain errors to status codes.
- `main.go` is only the composition root, environment configuration, embedded static assets, and sample seeding.

## Frontend

```text
src/
  ├── app/                 application controller and orchestration
  ├── features/
  │     ├── documents/     document contracts and API adapter
  │     ├── editor/        CodeMirror, Markdown parsing, Mermaid rendering
  │     ├── navigation/    application header and document navigation
  │     └── patches/       AI composer, review flow, and diffing
  ├── lib/                 framework-independent shared utilities
  └── ui/                  reusable domain-agnostic UI primitives
```

- Feature packages expose public entry points through `index.ts`; app code does not depend on their internal component paths.
- `ui` owns reusable primitives such as buttons, dialogs, and section labels without importing feature types.
- `app/useDocPatch` owns application state and async workflows while depending on feature-level APIs.
- `features/documents/api.ts` owns document transport details; `lib/http.ts` owns generic JSON transport.
- `features/editor/codeMirror.ts` and `MermaidPreview` isolate imperative third-party integrations behind Octane effects.
- Markdown parsing and patch diffing remain pure feature utilities.

## Core invariant

The model returns replacement text but never controls the replacement location. Applying a patch loads the authoritative SQLite document, verifies its version and original selected text, converts CodeMirror UTF-16 positions into UTF-8 byte boundaries, and replaces only that range inside one transaction.

## AI-assisted generation

```text
selected range + instruction
          │
          ├── ContextCompiler ──> context manifest (ancestors, references, backlinks)
          │
          └── Generator ────────> replacement Markdown
                                      │
                                      └── scoped Patch ──> review ──> transactional apply
```

The context manifest is deterministic, inspectable in the review dialog, persisted with the patch, and sent to the model as read-only material. The initial compiler intentionally avoids embeddings and model-based extraction. This keeps computation linear in document size and model usage limited to the final generation request.
