package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type WebhookPayload struct {
	EventType string         `json:"event_type"`
	Feed      WebhookFeed    `json:"feed"`
	Entries   []WebhookEntry `json:"entries"`
}

type WebhookFeed struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type WebhookEntry struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type EntryUpdater interface {
	UpdateEntryContent(entryID int64, content string) error
}

type WebhookHandler struct {
	Secret   string
	Curator  Curator
	Updater  EntryUpdater
	Profiler *Profiler
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Miniflux-Signature")
	if !h.validSignature(body, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-Miniflux-Event-Type")
	if eventType != "new_entries" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entries {
		h.processEntry(entry)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) validSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.Secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (h *WebhookHandler) processEntry(entry WebhookEntry) {
	profile := ""
	if h.Profiler != nil {
		profile = h.Profiler.Profile()
	}

	result, err := h.Curator.Curate(entry.Title, entry.Content, entry.URL, profile)
	if err != nil {
		log.Printf("curator error for entry %d: %v", entry.ID, err)
		return
	}

	summaryHTML := formatSummary(result)
	newContent := summaryHTML + entry.Content

	if err := h.Updater.UpdateEntryContent(entry.ID, newContent); err != nil {
		log.Printf("update error for entry %d: %v", entry.ID, err)
		return
	}

	log.Printf("processed entry %d [%s]: relevance=%d",
		entry.ID, entry.Title, result.Relevance)
}

func formatSummary(r *CurationResult) string {
	tags := strings.Join(r.Tags, ", ")
	return fmt.Sprintf(
		`<div style="border-left:3px solid #666;padding:8px 12px;margin-bottom:16px;font-size:14px;color:#999">`+
			`<strong>AI Summary:</strong> %s<br>`+
			`<strong>Tags:</strong> %s<br>`+
			`<strong>Relevance:</strong> %d/100`+
			`</div>`,
		r.Summary, tags, r.Relevance,
	)
}
