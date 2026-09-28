package main

import (
	"strings"
	"testing"
)

func TestRenderEmpty(t *testing.T) {
	html, err := RenderMarkdown([]byte{}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected valid HTML document")
	}
	if !strings.Contains(html, "<body>") {
		t.Error("expected <body> tag")
	}
}

func TestRenderTable(t *testing.T) {
	md := []byte("| A | B |\n|---|---|\n| 1 | 2 |")
	html, err := RenderMarkdown(md, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<table>") {
		t.Error("expected <table> tag")
	}
	if !strings.Contains(html, "<td>1</td>") {
		t.Error("expected table cell content")
	}
}

func TestRenderTaskList(t *testing.T) {
	md := []byte("- [x] Done\n- [ ] Todo")
	html, err := RenderMarkdown(md, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `type="checkbox"`) {
		t.Error("expected checkbox input")
	}
	if !strings.Contains(html, "checked") {
		t.Error("expected checked attribute")
	}
}

func TestCSPPresent(t *testing.T) {
	html, err := RenderMarkdown([]byte("# Hello"), "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Content-Security-Policy") {
		t.Error("expected CSP meta tag")
	}
	if !strings.Contains(html, "default-src 'none'") {
		t.Error("expected restrictive default-src")
	}
}

func TestBaseURISimple(t *testing.T) {
	uri, err := BaseURI("/tmp/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(uri, "file:///tmp/test/") {
		t.Errorf("expected file URI, got %s", uri)
	}
}

func TestBaseURISpecialChars(t *testing.T) {
	uri, err := BaseURI("/tmp/mein Ordner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(uri, "mein%20Ordner") {
		t.Errorf("expected escaped space, got %s", uri)
	}
}

func TestBaseURIUmlauts(t *testing.T) {
	uri, err := BaseURI("/tmp/Ärger")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(uri, "file:///tmp/") {
		t.Errorf("expected file URI, got %s", uri)
	}
}

func TestBaseURIHash(t *testing.T) {
	uri, err := BaseURI("/tmp/C# Notes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// # must be escaped so it's not treated as a fragment
	if strings.Contains(uri, "#") && !strings.Contains(uri, "%23") {
		t.Errorf("expected escaped #, got %s", uri)
	}
}
