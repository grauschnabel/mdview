# Security Policy

mdview opens untrusted Markdown files, so it is hardened by default:

- JavaScript markup in the document (`<script>`, event handlers, `javascript:`
  URLs) is disabled in WebKit itself.
- The page CSP (`default-src 'none'; script-src 'none'; base-uri 'none';
  form-action 'none'`) is a second layer; only inline styles and local (`file:`)
  or `data:` images are allowed.
- Exactly one navigation is allowed per programmatic load. Link clicks, meta
  refresh, redirects and new windows are ignored.
- The context menu is disabled and the network session is ephemeral.
- JavaScript evaluated through the WebKit API (used only to save and restore the
  scroll position) still runs, in an isolated script world.

## Supported versions

Only the latest release receives fixes. Versions up to 0.2.0 allowed a document
to navigate the window via `<meta http-equiv="refresh">`; please upgrade to 0.2.1.

## Reporting a vulnerability

Please report security issues privately via GitHub's
["Report a vulnerability"](https://github.com/grauschnabel/mdview/security/advisories/new)
form instead of opening a public issue. Expect a first response within a week.
