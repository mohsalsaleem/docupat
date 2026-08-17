package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docpatch/internal/document"
	"docpatch/internal/httpapi"
	"docpatch/internal/llm"
	"docpatch/internal/storage"
)

//go:embed dist/*
var web embed.FS

func main() {
	dataDir := env("DOCPATCH_DATA_DIR", "./data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatal(err)
	}
	repository, err := storage.Open(filepath.Join(dataDir, "docpatch.db"), newID, now)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	if err = repository.Seed(context.Background(), "Project Atlas — PRD", sample); err != nil {
		log.Fatal(err)
	}
	llamaURL := strings.TrimRight(env("LLAMA_BASE_URL", "http://127.0.0.1:8080"), "/")
	model := llm.NewClient(llamaURL, &http.Client{Timeout: 2 * time.Minute})
	service := document.NewService(repository, model, newID, now)
	dist, _ := fs.Sub(web, "dist")
	handler := httpapi.New(service, llamaURL).Router(dist)
	addr := "127.0.0.1:" + env("PORT", "4173")
	log.Printf("DocPatch %s · Llama %s", addr, llamaURL)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func newID() string {
	value := make([]byte, 12)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

const sample = `# Project Atlas — Product Requirements

## Problem

Engineering teams lose design context when requirements and architecture drift apart.

## Goals

- Make technical decisions easy to discover.
- Preserve a reviewable history of changes.

## System architecture

` + "```mermaid" + `
flowchart LR
    Browser --> API
    API --> SQLite
    API --> Llama[Local Llama]
` + "```" + `

## Authentication requirements

Users authenticate with email and password. Sessions expire after 24 hours.

### Constraints

- Credentials must never appear in logs.
- Authentication must remain available during regional failover.

## Open questions

- Which enterprise identity providers must be supported at launch?
`
