package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"

	"docpatch/internal/contextcompile"
	"docpatch/internal/document"
	"docpatch/internal/domain"
	"docpatch/internal/markdownindex"
	"docpatch/internal/storage"
)

func TestDocumentPatchWorkflow(t *testing.T) {
	handler, closeRepository := testHandler(t)
	defer closeRepository()

	health := perform(t, handler, http.MethodGet, "/api/health", nil)
	assertStatus(t, health, 200)
	var healthBody map[string]string
	decodeResponse(t, health, &healthBody)
	if healthBody["provider"] != "test" || healthBody["llm"] != "connected" {
		t.Fatalf("unexpected health: %#v", healthBody)
	}

	assertStatus(t, perform(t, handler, http.MethodGet, "/api/documents", nil), 200)
	createdResponse := perform(t, handler, http.MethodPost, "/api/documents", map[string]any{"title": "PRD", "content": "hello body"})
	assertStatus(t, createdResponse, 201)
	var created domain.Document
	decodeResponse(t, createdResponse, &created)

	gotResponse := perform(t, handler, http.MethodGet, "/api/documents/"+created.ID, nil)
	assertStatus(t, gotResponse, 200)
	var got domain.Document
	decodeResponse(t, gotResponse, &got)
	if got.Content != "hello body" || got.Version != 1 {
		t.Fatalf("unexpected document: %+v", got)
	}

	savedResponse := perform(t, handler, http.MethodPut, "/api/documents/"+created.ID, map[string]any{"title": "PRD v2", "content": "hello body", "version": 1})
	assertStatus(t, savedResponse, 200)
	var saved domain.Document
	decodeResponse(t, savedResponse, &saved)
	if saved.Version != 2 || saved.Title != "PRD v2" {
		t.Fatalf("unexpected saved document: %+v", saved)
	}

	proposalResponse := perform(t, handler, http.MethodPost, "/api/documents/"+created.ID+"/patches", map[string]any{"start": 6, "end": 10, "version": 2, "instruction": "rewrite", "useContext": true})
	assertStatus(t, proposalResponse, 201)
	var proposal domain.Patch
	decodeResponse(t, proposalResponse, &proposal)
	if proposal.Original != "body" || proposal.Replacement != "changed" {
		t.Fatalf("unexpected proposal: %+v", proposal)
	}

	patchesResponse := perform(t, handler, http.MethodGet, "/api/documents/"+created.ID+"/patches", nil)
	assertStatus(t, patchesResponse, 200)
	var patches []domain.Patch
	decodeResponse(t, patchesResponse, &patches)
	if len(patches) != 1 {
		t.Fatalf("patch count = %d", len(patches))
	}

	appliedResponse := perform(t, handler, http.MethodPost, "/api/patches/"+proposal.ID+"/apply", nil)
	assertStatus(t, appliedResponse, 200)
	var applied domain.Document
	decodeResponse(t, appliedResponse, &applied)
	if applied.Content != "hello changed" || applied.Version != 3 {
		t.Fatalf("unexpected applied document: %+v", applied)
	}

	assertStatus(t, perform(t, handler, http.MethodPost, "/api/patches/"+proposal.ID+"/reject", nil), 409)

	secondResponse := perform(t, handler, http.MethodPost, "/api/documents/"+created.ID+"/patches", map[string]any{"start": 6, "end": 13, "version": 3, "instruction": "rewrite"})
	assertStatus(t, secondResponse, 201)
	var second domain.Patch
	decodeResponse(t, secondResponse, &second)
	assertStatus(t, perform(t, handler, http.MethodPost, "/api/patches/"+second.ID+"/reject", nil), 200)
}

func TestValidationErrorsAndSPA(t *testing.T) {
	handler, closeRepository := testHandler(t)
	defer closeRepository()

	assertStatus(t, performRaw(handler, http.MethodPost, "/api/documents", []byte("{")), 400)
	assertStatus(t, perform(t, handler, http.MethodPost, "/api/documents", map[string]any{"title": " ", "content": "x"}), 400)
	assertStatus(t, perform(t, handler, http.MethodGet, "/api/documents/missing", nil), 404)
	assertStatus(t, perform(t, handler, http.MethodPut, "/api/documents/missing", map[string]any{"title": "x", "content": "x", "version": 1}), 409)

	index := perform(t, handler, http.MethodGet, "/unknown-route", nil)
	assertStatus(t, index, 200)
	if index.Body.String() != "<main>DocPatch</main>" {
		t.Fatalf("SPA body = %q", index.Body.String())
	}
}

type stubGenerator struct{}

func (stubGenerator) Generate(context.Context, document.GenerateInput) (string, error) {
	return "changed", nil
}
func (stubGenerator) Health(context.Context) string { return "connected" }

func testHandler(t *testing.T) (http.Handler, func()) {
	t.Helper()
	nextID := 0
	repository, err := storage.Open(filepath.Join(t.TempDir(), "test.db"), func() string {
		nextID++
		return "id-" + string(rune('a'+nextID))
	}, func() string { return "2026-01-01T00:00:00Z" })
	if err != nil {
		t.Fatal(err)
	}
	service := document.NewService(repository, stubGenerator{}, contextcompile.New(repository, 6000), markdownindex.New(), func() string {
		nextID++
		return "patch-" + string(rune('a'+nextID))
	}, func() string { return "2026-01-01T00:00:00Z" })
	static := fstest.MapFS{"index.html": {Data: []byte("<main>DocPatch</main>"), Mode: fs.FileMode(0o644)}}
	return New(service, LLMInfo{Provider: "test", Model: "test-model", BaseURL: "local"}).Router(static), func() { _ = repository.Close() }
}

func perform(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	return performRaw(handler, method, path, payload)
}

func performRaw(handler http.Handler, method, path string, payload []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if response.Code != expected {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, expected, response.Body.String())
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}
