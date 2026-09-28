# Security Policy

mdview opens untrusted Markdown files, so it is hardened by default:

- The page CSP is `default-src 'none'; script-src 'none'`; only inline styles and
  local (`file:`) or `data:` images are allowed.
- All user-initiated navigation (link clicks) is blocked.
- JavaScript is enabled in WebKit only so the application itself can restore the
  scroll position; scripts inside the document never execute.

## Reporting a vulnerability

Please report security issues privately via GitHub's
["Report a vulnerability"](https://github.com/grauschnabel/mdview/security/advisories/new)
form instead of opening a public issue. Expect a first response within a week.
