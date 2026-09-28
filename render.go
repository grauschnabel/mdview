// This file turns Markdown into a self-contained HTML document and computes the
// base URI used to resolve relative resources such as images.
package main

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// cssStyle is the built-in stylesheet, inlined into every page. It is inline
// because the CSP forbids loading external stylesheets.
const cssStyle = `
body {
	max-width: 48em;
	margin: 2em auto;
	padding: 0 1em;
	font-family: system-ui, -apple-system, sans-serif;
	font-size: 16px;
	line-height: 1.6;
	color: #1a1a1a;
	background: #fff;
}

h1, h2, h3, h4, h5, h6 {
	margin-top: 1.5em;
	margin-bottom: 0.5em;
	line-height: 1.3;
}

h1 { font-size: 2em; border-bottom: 1px solid #ddd; padding-bottom: 0.3em; }
h2 { font-size: 1.5em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }

a { color: #0366d6; text-decoration: none; }

code {
	background: #f4f4f4;
	padding: 0.2em 0.4em;
	border-radius: 3px;
	font-size: 0.9em;
}

pre {
	background: #f4f4f4;
	padding: 1em;
	border-radius: 6px;
	overflow-x: auto;
	line-height: 1.4;
}

pre code {
	background: none;
	padding: 0;
}

blockquote {
	border-left: 4px solid #ddd;
	margin: 1em 0;
	padding: 0.5em 1em;
	color: #555;
}

table {
	border-collapse: collapse;
	width: 100%;
	margin: 1em 0;
}

th, td {
	border: 1px solid #ddd;
	padding: 0.5em 0.75em;
	text-align: left;
}

th { background: #f4f4f4; font-weight: 600; }
tr:nth-child(even) { background: #fafafa; }

img { max-width: 100%; height: auto; }

hr { border: none; border-top: 1px solid #ddd; margin: 2em 0; }

ul, ol { padding-left: 2em; }
li { margin: 0.25em 0; }

.footnotes {
	font-size: 0.9em;
	color: #555;
}

dt { font-weight: 600; margin-top: 0.75em; }
dd { margin-left: 2em; }

input[type="checkbox"] {
	margin-right: 0.5em;
}
`

// csp is the page's Content-Security-Policy. Note that base-uri and form-action
// do not fall back to default-src and must be set explicitly.
const csp = "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; " +
	"img-src file: data:; base-uri 'none'; form-action 'none'"

// maxHighlightSize is the largest document that is syntax highlighted. Chroma
// can take seconds on pathological megabyte-sized code blocks, and the first
// render happens before the window opens, so bigger files are shown unhighlighted.
const maxHighlightSize = 256 * 1024

// highlightStyle is the Chroma theme. Highlighting uses CSS classes (see
// highlightCSS) rather than inline styles, which keeps the HTML small and lets
// code blocks share the page's own background.
const highlightStyle = "github"

// highlightFormatter is the Chroma HTML formatter shared by the parser and
// the stylesheet generator so both agree on class names.
var highlightFormatter = chromahtml.New(chromahtml.WithClasses(true))

// highlightCSS holds the token classes for highlightStyle, generated once.
var highlightCSS = func() string {
	var buf bytes.Buffer
	if err := highlightFormatter.WriteCSS(&buf, styles.Get(highlightStyle)); err != nil {
		return ""
	}
	return buf.String()
}()

// newParser builds a goldmark instance. Raw HTML is passed through on purpose;
// safety relies on the CSP and on the viewer's navigation policy, not on
// sanitising.
func newParser(highlight bool) goldmark.Markdown {
	exts := []goldmark.Extender{
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.Footnote,
		extension.DefinitionList,
	}
	if highlight {
		exts = append(exts, highlighting.NewHighlighting(
			highlighting.WithStyle(highlightStyle),
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		))
	}
	return goldmark.New(
		goldmark.WithExtensions(exts...),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

// md renders normal documents; mdPlain renders very large ones without
// syntax highlighting.
var (
	md      = newParser(true)
	mdPlain = newParser(false)
)

// RenderMarkdown converts Markdown bytes to a complete HTML document.
// The Content-Security-Policy meta tag is emitted first, before any
// user-controlled content, so the policy is in force while the document is parsed.
func RenderMarkdown(mdBytes []byte) (string, error) {
	parser := md
	if len(mdBytes) > maxHighlightSize {
		parser = mdPlain
	}
	var buf bytes.Buffer
	if err := parser.Convert(mdBytes, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	doc := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="%s">
<style>%s%s</style>
</head>
<body>
%s
</body>
</html>`, csp, cssStyle, highlightCSS, buf.String())

	return doc, nil
}

// BaseURI constructs a file:// URI from a directory path, properly escaping
// special characters (spaces, umlauts, #, etc.).
func BaseURI(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	// Ensure trailing slash so relative paths resolve correctly.
	if absDir[len(absDir)-1] != '/' {
		absDir += "/"
	}

	u := &url.URL{
		Scheme: "file",
		Path:   absDir,
	}
	return u.String(), nil
}
