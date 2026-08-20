# Stellarity

An AI-assisted technical-document workspace for precise, reviewable Markdown changes, Mermaid diagrams, and durable SQLite history. Stellarity compiles focused context from document structure and explicit links, while the Go backend only permits the model to replace the explicitly selected range.

<img src="./docs/assets/docpatch-demo.gif" alt="Stellarity focused editing workflow" width="100%">

## Run

Requirements: Go 1.26+, Node.js 22.22.2+, and pnpm. The default model provider is an OpenAI-compatible local `llama-server`.

```bash
pnpm install
pnpm build
go run .
```

Open <http://127.0.0.1:4173>. By default the app connects to `http://127.0.0.1:8080`.

```bash
LLM_BASE_URL=http://127.0.0.1:8080 LLM_MODEL=local-model PORT=4173 go run .
```

Click a heading in the outline or highlight Markdown, write an instruction, and review the proposed diff before applying it.

For frontend development, run the Go API and `pnpm dev` in separate terminals. Vite runs on port 5173 and proxies `/api` to Go on port 4173.

## Model configuration

Stellarity supports OpenAI-compatible APIs—including local `llama-server`—and Anthropic. Configuration is environment-based:

| Variable | Description |
| --- | --- |
| `LLM_PROVIDER` | `openai` (default) or `anthropic` |
| `LLM_MODEL` | Provider model identifier; defaults to `local-model` for OpenAI-compatible APIs |
| `LLM_BASE_URL` | API base URL; defaults to local Llama for `openai` and Anthropic's API for `anthropic` |
| `LLM_API_KEY` | Provider API key, when required |
| `SEMANTIC_CONTEXT_ENABLED` | Enables cached semantic retrieval only when explicit context is insufficient; defaults to `false` |
| `EMBEDDING_MODEL` | Model used by the optional OpenAI-compatible embeddings endpoint |
| `EMBEDDING_BASE_URL` | Embeddings API base URL; defaults to `LLM_BASE_URL` |
| `EMBEDDING_API_KEY` | Embeddings API key; defaults to the configured model API key |

`OPENAI_API_KEY` and `ANTHROPIC_API_KEY` are supported as provider-specific alternatives to `LLM_API_KEY`. The legacy `LLAMA_BASE_URL` variable remains supported.

```bash
# OpenAI
LLM_PROVIDER=openai LLM_MODEL=your-openai-model OPENAI_API_KEY=... go run .

# Anthropic
LLM_PROVIDER=anthropic LLM_MODEL=your-anthropic-model ANTHROPIC_API_KEY=... go run .
```

## Verify

```bash
go test ./...
pnpm build
pnpm coverage
```

`pnpm coverage` runs the internal backend suite with cross-package instrumentation and fails when statement coverage drops below 80%. The current suite covers document workflows through the HTTP boundary, real SQLite persistence, scoped-patch invariants, and both model-provider adapters.

Documents, immutable revisions, section/link indexes, patch proposals, and their context manifests are stored in `data/docpatch.db`. When focused context is enabled, Stellarity deterministically resolves ancestor headings, Markdown links, wiki links, and backlinks across the workspace within a 6,000-character budget. Only those sources and the selected target are sent to the configured provider. Context compilation itself uses no model tokens. With the default local configuration, no document data is sent to a hosted service.

Semantic retrieval is an opt-in fallback. When enabled, it runs only if structural and explicit relationships contribute fewer than 800 context characters. Section embeddings are cached by content hash and model in SQLite; unchanged sections are not embedded again. Matches below the similarity threshold are discarded, at most three semantic sources are admitted, and all sources still share the same 6,000-character budget.

Every generated patch includes a rule-based confidence assessment. Stellarity scores structural ancestry, explicit relationships, semantic support, and unresolved links without making another model call. The review dialog shows the score and evidence counts so thinly grounded edits are visible before they are applied.

See [docs/architecture.md](docs/architecture.md) for package boundaries, dependency direction, and extension points.
