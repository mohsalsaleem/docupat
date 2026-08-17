# DocPatch

Code-editor-style AI patches for technical Markdown documents, with Mermaid diagrams and durable SQLite history. The model may read document context, but the Go backend only permits it to replace the explicitly selected range.

## Run

Requirements: Go 1.26+, Node.js 22.22.2+, pnpm, and an OpenAI-compatible `llama-server`.

```bash
pnpm install
pnpm build
go run .
```

Open <http://127.0.0.1:4173>. By default the app connects to `http://127.0.0.1:8080`.

```bash
LLAMA_BASE_URL=http://127.0.0.1:8080 PORT=4173 go run .
```

Click a heading in the outline or highlight Markdown, write an instruction, and review the proposed diff before applying it.

For frontend development, run the Go API and `pnpm dev` in separate terminals. Vite runs on port 5173 and proxies `/api` to Go on port 4173.

## Verify

```bash
go test ./...
pnpm build
```

Documents, immutable revisions, and patch proposals are stored in `data/docpatch.db`. It sends the selected text and, when enabled, the rest of the document as read-only context to the local model. No document data is sent to a hosted service by the app.

See [docs/architecture.md](docs/architecture.md) for package boundaries, dependency direction, and extension points.
