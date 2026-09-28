#ifndef MDVIEW_VIEWER_H
#define MDVIEW_VIEWER_H

// viewer_run opens the window and blocks until it is closed.
// Returns 0 on success, non-zero if GTK could not be initialised.
int viewer_run(const char *title, const char *html, const char *base_uri);

// viewer_reload replaces the displayed HTML and keeps the scroll position.
// Safe to call from any thread.
void viewer_reload(const char *html, const char *base_uri);

#endif
