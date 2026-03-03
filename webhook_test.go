package main

import (
	"bytes"
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

func (m *mockCurator) Curate(title, content, url, userProfile string) (*CurationResult, error) {
	return m.result, m.err
}

type mockUpdater struct {
	updated map[int64]string
}

func (m *mockUpdater) UpdateEntryContent(entryID int64, content string) error {
	m.updated[entryID] = content
	return nil
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
		Summary: "Test summary", Tags: []string{"go"}, Relevance: 90, Reason: "Relevant",
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

func TestWebhook_LowRelevance_StillUpdated(t *testing.T) {
	mc := &mockCurator{result: &CurationResult{
		Summary: "Meh", Tags: []string{"misc"}, Relevance: 30, Reason: "Not relevant",
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

	if _, ok := mu.updated[2]; !ok {
		t.Error("entry 2 should still be updated with summary")
	}
}
