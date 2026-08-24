package httpapi

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"docpatch/internal/document"
	"docpatch/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	service *document.Service
	llm     LLMInfo
}

type LLMInfo struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseUrl"`
}

func New(service *document.Service, info LLMInfo) *Handler {
	return &Handler{service: service, llm: info}
}

func (h *Handler) Router(static fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(3*time.Minute))
	r.Get("/api/health", h.health)
	r.Get("/api/documents", h.listDocuments)
	r.Post("/api/documents", h.createDocument)
	r.Route("/api/documents/{id}", func(r chi.Router) {
		r.Get("/", h.getDocument)
		r.Put("/", h.saveDocument)
		r.Delete("/", h.deleteDocument)
		r.Post("/restore", h.restoreDocument)
		r.Get("/patches", h.listPatches)
		r.Post("/patches", h.proposePatch)
	})
	r.Post("/api/patches/{id}/apply", h.applyPatch)
	r.Post("/api/patches/{id}/reject", h.rejectPatch)
	r.Handle("/*", spaHandler(http.FS(static)))
	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func decode(r *http.Request, value any) error {
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 2<<20)).Decode(value)
}
func fail(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := err.Error()
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = 404
		message = "resource not found"
	case errors.Is(err, domain.ErrConflict):
		status = 409
		message = "resource changed or operation is no longer pending"
	case errors.Is(err, domain.ErrInvalid):
		status = 400
		message = "invalid request"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"app": "ok", "llm": h.service.LLMHealth(r.Context()), "provider": h.llm.Provider, "model": h.llm.Model, "baseUrl": h.llm.BaseURL})
}
func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) createDocument(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, domain.ErrInvalid)
		return
	}
	item, err := h.service.Create(r.Context(), in.Title, in.Content)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 201, item)
}
func (h *Handler) saveDocument(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Version int    `json:"version"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, domain.ErrInvalid)
		return
	}
	item, err := h.service.Save(r.Context(), chi.URLParam(r, "id"), in.Version, in.Title, in.Content)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Delete(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) restoreDocument(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Version int `json:"version"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, domain.ErrInvalid)
		return
	}
	item, err := h.service.Restore(r.Context(), chi.URLParam(r, "id"), in.Version)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) listPatches(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Patches(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (h *Handler) proposePatch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Start       int    `json:"start"`
		End         int    `json:"end"`
		Version     int    `json:"version"`
		Instruction string `json:"instruction"`
		UseContext  bool   `json:"useContext"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, domain.ErrInvalid)
		return
	}
	item, err := h.service.Propose(r.Context(), document.ProposeInput{DocumentID: chi.URLParam(r, "id"), Version: in.Version, Selection: domain.Selection{Start: in.Start, End: in.End}, Instruction: in.Instruction, UseContext: in.UseContext})
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 201, item)
}
func (h *Handler) applyPatch(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Apply(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) rejectPatch(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Reject(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "rejected"})
}

func spaHandler(files http.FileSystem) http.Handler {
	server := http.FileServer(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := files.Open(path); err == nil {
			f.Close()
			server.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		server.ServeHTTP(w, r)
	})
}
