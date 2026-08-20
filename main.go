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

	"docpatch/internal/contextcompile"
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
	modelConfig := loadLLMConfig()
	model, err := llm.NewClient(modelConfig, &http.Client{Timeout: 2 * time.Minute})
	if err != nil {
		log.Fatal(err)
	}
	service := document.NewService(repository, model, contextcompile.New(6000), newID, now)
	dist, _ := fs.Sub(web, "dist")
	handler := httpapi.New(service, httpapi.LLMInfo{Provider: modelConfig.Provider, Model: modelConfig.Model, BaseURL: modelConfig.BaseURL}).Router(dist)
	addr := "127.0.0.1:" + env("PORT", "4173")
	log.Printf("Stellarity %s · %s/%s · %s", addr, modelConfig.Provider, modelConfig.Model, modelConfig.BaseURL)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func loadLLMConfig() llm.Config {
	provider := strings.ToLower(env("LLM_PROVIDER", "openai"))
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("LLAMA_BASE_URL") // Backward compatibility.
	}
	if baseURL == "" {
		if provider == "anthropic" {
			baseURL = "https://api.anthropic.com"
		} else {
			baseURL = "http://127.0.0.1:8080"
		}
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" && provider == "openai" {
		model = "local-model"
	}
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		if provider == "anthropic" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		} else {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}
	return llm.Config{Provider: provider, BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, Model: model}
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
    API --> Model[Configured model]
` + "```" + `

## Authentication requirements

Users authenticate with email and password. Sessions expire after 24 hours.

### Constraints

- Credentials must never appear in logs.
- Authentication must remain available during regional failover.

## Open questions

- Which enterprise identity providers must be supported at launch?
`
