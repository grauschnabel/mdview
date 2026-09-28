package main

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

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

input[type="checkbox"] {
	margin-right: 0.5em;
}
`

// RenderMarkdown converts Markdown bytes to a complete HTML document.
// baseDir is used to construct the base URI for resolving relative paths (e.g. images).
func RenderMarkdown(mdBytes []byte, baseDir string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(mdBytes, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src file: data:;">
<style>%s</style>
</head>
<body>
%s
</body>
</html>`, cssStyle, buf.String())

	return html, nil
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
