package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockCurator struct {
	result *CurationResult
	err    error
}

func (m *mockCurator) Curate(title, content, url string) (*CurationResult, error) {
	return m.result, m.err
}

type mockUpdater struct {
	updated map[int64]string
}

func (m *mockUpdater) UpdateEntryContent(entryID int64, content string) error {
	m.updated[entryID] = content
	return nil
}

type mockEmbedder struct {
	embedding []float32
	err       error
}

func (m *mockEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return m.embedding, m.err
}

func sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestWebhook_InvalidSignature(t *testing.T) {
	h := &WebhookHandler{Secret: "test-secret"}
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Miniflux-Signature", "bad-sig")
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestWebhook_ValidSignature_WrongEvent(t *testing.T) {
	h := &WebhookHandler{Secret: "test-secret"}
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Miniflux-Signature", sign(body, "test-secret"))
	req.Header.Set("X-Miniflux-Event-Type", "other_event")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWebhook_ProcessesEntries(t *testing.T) {
	mc := &mockCurator{result: &CurationResult{
		Summary: "Test summary", Tags: []string{"go"}, Reason: "Relevant",
	}}
	mu := &mockUpdater{updated: make(map[int64]string)}
	h := &WebhookHandler{Secret: "s", Curator: mc, Updater: mu}

	payload := WebhookPayload{
		EventType: "new_entries",
		Entries: []WebhookEntry{
			{ID: 1, Title: "Test", URL: "https://example.com", Content: "<p>Original</p>"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Miniflux-Signature", sign(body, "s"))
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if _, ok := mu.updated[1]; !ok {
		t.Error("entry 1 not updated")
	}
}

func TestWebhook_NoEmbedder_FallbackRelevance(t *testing.T) {
	mc := &mockCurator{result: &CurationResult{
		Summary: "Test", Tags: []string{"misc"}, Reason: "ok",
	}}
	mu := &mockUpdater{updated: make(map[int64]string)}
	h := &WebhookHandler{Secret: "s", Curator: mc, Updater: mu}

	payload := WebhookPayload{
		EventType: "new_entries",
		Entries:   []WebhookEntry{{ID: 2, Title: "Boring", Content: "stuff"}},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Miniflux-Signature", sign(body, "s"))
	req.Header.Set("X-Miniflux-Event-Type", "new_entries")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	content := mu.updated[2]
	if content == "" {
		t.Error("entry 2 should be updated")
	}
	if !bytes.Contains([]byte(content), []byte("50/100")) {
		t.Errorf("expected fallback relevance 50, got content: %s", content)
	}
}
