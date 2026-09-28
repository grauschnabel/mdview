package main

import (
	"strings"
	"testing"
)

// TestRenderEmpty checks that empty input still yields a complete HTML document.
func TestRenderEmpty(t *testing.T) {
	html, err := RenderMarkdown([]byte{})
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

// TestRenderTable checks that the GFM table extension is enabled.
func TestRenderTable(t *testing.T) {
	md := []byte("| A | B |\n|---|---|\n| 1 | 2 |")
	html, err := RenderMarkdown(md)
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

// TestRenderTaskList checks that the task list extension is enabled.
func TestRenderTaskList(t *testing.T) {
	md := []byte("- [x] Done\n- [ ] Todo")
	html, err := RenderMarkdown(md)
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

// TestCSPExact checks that the CSP meta tag is emitted with the expected policy.
func TestCSPExact(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# Hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `<meta http-equiv="Content-Security-Policy" content="` + csp + `">`
	if !strings.Contains(doc, want) {
		t.Errorf("expected CSP meta tag %q", want)
	}
	for _, directive := range []string{
		"default-src 'none'", "script-src 'none'", "base-uri 'none'", "form-action 'none'",
	} {
		if !strings.Contains(csp, directive) {
			t.Errorf("CSP is missing %q", directive)
		}
	}
}

// The CSP must precede every byte of user-controlled content, otherwise a
// document could influence how it is parsed before the policy applies.
func TestCSPBeforeUserContent(t *testing.T) {
	doc, err := RenderMarkdown([]byte("MARKER"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Index(doc, "Content-Security-Policy") > strings.Index(doc, "MARKER") {
		t.Error("CSP meta tag must come before document content")
	}
}

// Raw HTML is passed through on purpose. This test documents that safety comes
// from the CSP and the viewer's navigation policy (viewer.c), not from sanitising.
func TestRawHTMLPassesThrough(t *testing.T) {
	for _, raw := range []string{
		`<script>alert(1)</script>`,
		`<img src=x onerror="alert(1)">`,
		`<meta http-equiv="refresh" content="0;url=https://example.com/">`,
	} {
		doc, err := RenderMarkdown([]byte(raw))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(doc, raw) {
			t.Errorf("expected raw HTML %q to be kept", raw)
		}
	}
}

// NUL bytes must not reach C.CString, where they would truncate the document.
func TestNoNULBytes(t *testing.T) {
	doc, err := RenderMarkdown([]byte("before\x00after"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.ContainsRune(doc, 0) {
		t.Error("rendered document contains a NUL byte")
	}
}

// TestBaseURI checks that special characters in directory names are escaped.
func TestBaseURI(t *testing.T) {
	tests := []struct{ dir, want string }{
		{"/tmp/test", "file:///tmp/test/"},
		{"/tmp/mein Ordner", "file:///tmp/mein%20Ordner/"},
		{"/tmp/Ärger", "file:///tmp/%C3%84rger/"},
		{"/tmp/C# Notes", "file:///tmp/C%23%20Notes/"},
		{"/tmp/100%", "file:///tmp/100%25/"},
		{"/tmp/a?b", "file:///tmp/a%3Fb/"},
		{"/", "file:///"},
	}
	for _, tt := range tests {
		got, err := BaseURI(tt.dir)
		if err != nil {
			t.Fatalf("BaseURI(%q): %v", tt.dir, err)
		}
		if got != tt.want {
			t.Errorf("BaseURI(%q) = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

// TestFootnotes checks that the footnote extension is enabled.
func TestFootnotes(t *testing.T) {
	doc, err := RenderMarkdown([]byte("Text[^1]\n\n[^1]: Note."))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc, `class="footnotes"`) {
		t.Error("expected footnotes section")
	}
}

// TestDefinitionList checks that the definition list extension is enabled.
func TestDefinitionList(t *testing.T) {
	doc, err := RenderMarkdown([]byte("Term\n: Definition"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc, "<dl>") || !strings.Contains(doc, "<dd>") {
		t.Error("expected definition list")
	}
}

// TestSyntaxHighlighting checks that fenced code is tokenised into classed
// spans and that the token stylesheet is included.
func TestSyntaxHighlighting(t *testing.T) {
	doc, err := RenderMarkdown([]byte("```go\nfunc main() {}\n```"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc, `<span class="kd">func</span>`) {
		t.Errorf("expected tokenised keyword, got:\n%s", doc)
	}
	if !strings.Contains(doc, ".chroma") {
		t.Error("expected Chroma stylesheet")
	}
}

// TestLargeDocumentSkipsHighlighting guards against Chroma slowness on huge
// code blocks.
func TestLargeDocumentSkipsHighlighting(t *testing.T) {
	big := "```go\nfunc main() {}\n```\n" + strings.Repeat("x", maxHighlightSize)
	doc, err := RenderMarkdown([]byte(big))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(doc, `<span class="kd">`) {
		t.Error("large documents must not be highlighted")
	}
}

// TestHostileCodeFence checks that a malicious info string and code body stay
// inert text, for known and unknown languages.
func TestHostileCodeFence(t *testing.T) {
	for _, src := range []string{
		"```\"><script>alert(1)</script>\nx\n```",
		"```nosuchlang\n<script>alert(1)</script>\n```",
		"```html\n<img src=x onerror=alert(1)>\n```",
	} {
		doc, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		body := doc[strings.Index(doc, "<body>"):]
		if strings.Contains(body, "<script>") || strings.Contains(body, "<img src=x") {
			t.Errorf("code fence leaked live HTML for %q", src)
		}
	}
}

// TestHostileFootnoteLabel checks that footnote labels never reach ids or hrefs.
func TestHostileFootnoteLabel(t *testing.T) {
	doc, err := RenderMarkdown([]byte("A[^javascript:alert(1)]\n\n[^javascript:alert(1)]: note"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(doc, `href="javascript`) || !strings.Contains(doc, `href="#fn:1"`) {
		t.Error("footnote label must be replaced by a numeric id")
	}
}
