package main

import (
	"testing"
)

func TestParseCurationResult_Valid(t *testing.T) {
	input := `{"summary":"A good article","tags":["go","ai"],"relevance":85,"reason":"Relevant to AI"}`
	r, err := parseCurationResult(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Summary != "A good article" {
		t.Errorf("Summary = %q", r.Summary)
	}
	if len(r.Tags) != 2 || r.Tags[0] != "go" {
		t.Errorf("Tags = %v", r.Tags)
	}
	if r.Relevance != 85 {
		t.Errorf("Relevance = %d", r.Relevance)
	}
}

func TestParseCurationResult_MarkdownFences(t *testing.T) {
	input := "```json\n{\"summary\":\"test\",\"tags\":[],\"relevance\":50,\"reason\":\"ok\"}\n```"
	r, err := parseCurationResult(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Summary != "test" {
		t.Errorf("Summary = %q", r.Summary)
	}
}

func TestParseCurationResult_Invalid(t *testing.T) {
	_, err := parseCurationResult("not json")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCurationResult_ClampRelevance(t *testing.T) {
	input := `{"summary":"x","tags":[],"relevance":150,"reason":"x"}`
	r, err := parseCurationResult(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Relevance != 100 {
		t.Errorf("Relevance = %d, want 100", r.Relevance)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 100); got != "short" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("abcdefgh", 5); got != "abcde..." {
		t.Errorf("truncate long = %q", got)
	}
}
