/*
 * viewer.c - the GTK 4 window and WebKitGTK web view of mdview.
 *
 * Besides showing the rendered document this file enforces the security model
 * for untrusted Markdown: no JavaScript from the document, exactly one
 * navigation per programmatic load, no new windows, no context menu and an
 * ephemeral (non-persistent) network session.
 */
#include <string.h>

#include <gtk/gtk.h>
#include <webkit/webkit.h>

#include "viewer.h"

// Scripts of our own run in a separate JavaScript world so a hostile page
// cannot tamper with what they read.
#define SCRIPT_WORLD "mdview"

// Main loop that keeps viewer_run() blocked until the window is closed.
static GMainLoop *loop;
// The single web view. Reset to NULL when destroyed so late reload requests
// from the file watcher become no-ops instead of touching a dead widget.
static WebKitWebView *webview;

// Scroll position to restore once the next load has finished.
static double pending_scroll = 0;

// Set immediately before every programmatic load. decide_policy lets exactly
// one navigation through per load; everything else (meta refresh, redirects,
// link clicks, new windows) is ignored.
static gboolean allow_next_load = FALSE;

// load_document loads html into the web view, allowing exactly the one
// navigation this load causes (see allow_next_load).
static void load_document(const char *html, const char *base_uri) {
	allow_next_load = TRUE;
	webkit_web_view_load_html(webview, html, base_uri);
}

// on_load_changed restores the scroll position after a reload.
static void on_load_changed(WebKitWebView *wv, WebKitLoadEvent event, gpointer user_data) {
	if (event == WEBKIT_LOAD_FINISHED && pending_scroll > 0) {
		char js[64];
		g_snprintf(js, sizeof js, "window.scrollTo(0, %d);", (int)pending_scroll);
		webkit_web_view_evaluate_javascript(wv, js, -1, SCRIPT_WORLD, NULL, NULL, NULL, NULL);
		pending_scroll = 0;
	}
}

// is_same_document_anchor reports whether a link click only jumps to a fragment
// (#...) inside the page that is currently shown, e.g. a footnote link.
static gboolean is_same_document_anchor(WebKitWebView *wv, WebKitPolicyDecision *decision) {
	WebKitNavigationAction *action =
	    webkit_navigation_policy_decision_get_navigation_action(WEBKIT_NAVIGATION_POLICY_DECISION(decision));
	if (webkit_navigation_action_get_navigation_type(action) != WEBKIT_NAVIGATION_TYPE_LINK_CLICKED) {
		return FALSE;
	}
	const char *target = webkit_uri_request_get_uri(webkit_navigation_action_get_request(action));
	const char *current = webkit_web_view_get_uri(wv);
	const char *hash = target ? strchr(target, '#') : NULL;
	if (!hash || !current) {
		return FALSE;
	}
	size_t prefix = (size_t)(hash - target);
	const char *current_hash = strchr(current, '#');
	size_t current_len = current_hash ? (size_t)(current_hash - current) : strlen(current);
	return prefix == current_len && strncmp(target, current, prefix) == 0;
}

// decide_policy is the navigation gatekeeper: it allows the navigation caused by
// load_document and in-page anchor links, and ignores everything else the page
// tries to do (meta refresh, redirects, external links, new windows).
static gboolean decide_policy(WebKitWebView *wv, WebKitPolicyDecision *decision,
                              WebKitPolicyDecisionType type, gpointer user_data) {
	switch (type) {
	case WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION:
		if (allow_next_load) {
			allow_next_load = FALSE;
			return FALSE;
		}
		if (is_same_document_anchor(wv, decision)) {
			return FALSE;
		}
		/* fall through */
	case WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION:
		webkit_policy_decision_ignore(decision);
		return TRUE;
	default:
		return FALSE;
	}
}

// on_context_menu disables the context menu: its "Download ..." entries would
// start downloads without going through decide_policy.
static gboolean on_context_menu(WebKitWebView *wv, WebKitContextMenu *menu,
                                WebKitHitTestResult *hit, gpointer user_data) {
	return TRUE;
}

// on_key_pressed: q, Ctrl+Q and Ctrl+W quit.
static gboolean on_key_pressed(GtkEventControllerKey *ctrl, guint keyval, guint keycode,
                               GdkModifierType state, gpointer user_data) {
	state &= gtk_accelerator_get_default_mod_mask();
	state &= ~GDK_LOCK_MASK; // Caps Lock must not disable the shortcuts
	guint key = gdk_keyval_to_lower(keyval);
	if ((key == GDK_KEY_q && (state == 0 || state == GDK_CONTROL_MASK)) ||
	    (key == GDK_KEY_w && state == GDK_CONTROL_MASK)) {
		g_main_loop_quit(loop);
		return TRUE;
	}
	return FALSE;
}

// on_close_request ends the main loop when the window is closed.
static gboolean on_close_request(GtkWindow *win, gpointer user_data) {
	g_main_loop_quit(loop);
	return FALSE;
}

// on_webview_destroy forgets the web view so pending reloads are dropped.
static void on_webview_destroy(GtkWidget *widget, gpointer user_data) {
	webview = NULL;
}

// viewer_run builds the window, shows the document and runs the main loop.
int viewer_run(const char *title, const char *html, const char *base_uri) {
	if (!gtk_init_check()) {
		return 1;
	}

	GtkWidget *window = gtk_window_new();
	gtk_window_set_title(GTK_WINDOW(window), title);
	gtk_window_set_default_size(GTK_WINDOW(window), 900, 700);
	g_signal_connect(window, "close-request", G_CALLBACK(on_close_request), NULL);

	// Ephemeral session: nothing (cache, cookies, storage) is written to disk.
	WebKitNetworkSession *session = webkit_network_session_new_ephemeral();
	webview = WEBKIT_WEB_VIEW(g_object_new(WEBKIT_TYPE_WEB_VIEW, "network-session", session, NULL));
	g_object_unref(session);

	// Hardened settings. JavaScript that comes from the document is switched off
	// entirely (<script>, event handlers, javascript: URLs); the CSP is a second
	// layer. API-evaluated scripts still work, which the app needs to read and
	// restore the scroll position.
	WebKitSettings *settings = webkit_web_view_get_settings(webview);
	webkit_settings_set_enable_javascript_markup(settings, FALSE);
	g_signal_connect(webview, "decide-policy", G_CALLBACK(decide_policy), NULL);
	g_signal_connect(webview, "context-menu", G_CALLBACK(on_context_menu), NULL);
	g_signal_connect(webview, "load-changed", G_CALLBACK(on_load_changed), NULL);
	g_signal_connect(webview, "destroy", G_CALLBACK(on_webview_destroy), NULL);
	gtk_window_set_child(GTK_WINDOW(window), GTK_WIDGET(webview));

	GtkEventController *keys = gtk_event_controller_key_new();
	gtk_event_controller_set_propagation_phase(keys, GTK_PHASE_CAPTURE);
	g_signal_connect(keys, "key-pressed", G_CALLBACK(on_key_pressed), NULL);
	gtk_widget_add_controller(window, keys);

	load_document(html, base_uri);
	gtk_window_present(GTK_WINDOW(window));

	loop = g_main_loop_new(NULL, FALSE);
	g_main_loop_run(loop);
	g_main_loop_unref(loop);
	return 0;
}

// reload_req carries a reload request from the watcher thread to the GTK thread.
// It owns copies of both strings.
typedef struct {
	char *html;
	char *base_uri;
} reload_req;

// free_req releases a reload_req and its strings.
static void free_req(reload_req *req) {
	g_free(req->html);
	g_free(req->base_uri);
	g_free(req);
}

// on_scroll_read receives the current scroll offset, then loads the new document.
static void on_scroll_read(GObject *source, GAsyncResult *res, gpointer user_data) {
	reload_req *req = user_data;
	JSCValue *r = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(source), res, NULL);
	if (r) {
		double v = jsc_value_to_double(r);
		g_object_unref(r);
		// Also rejects NaN and absurd values before they are cast to int.
		pending_scroll = (v > 0 && v < 1e7) ? v : 0;
	}
	if (webview) {
		load_document(req->html, req->base_uri);
	}
	free_req(req);
}

// do_reload runs on the GTK main thread.
static gboolean do_reload(gpointer data) {
	reload_req *req = data;
	if (!webview) {
		free_req(req);
		return G_SOURCE_REMOVE;
	}
	// While a load is still in flight the page is at the top; reading scrollY now
	// would overwrite the position saved by the previous reload.
	if (webkit_web_view_is_loading(webview)) {
		load_document(req->html, req->base_uri);
		free_req(req);
		return G_SOURCE_REMOVE;
	}
	webkit_web_view_evaluate_javascript(webview, "window.scrollY;", -1, SCRIPT_WORLD, NULL, NULL,
	                                    on_scroll_read, req);
	return G_SOURCE_REMOVE;
}

// viewer_reload copies its arguments and schedules the reload on the GTK main
// loop, which makes it safe to call from any thread.
void viewer_reload(const char *html, const char *base_uri) {
	reload_req *req = g_new0(reload_req, 1);
	req->html = g_strdup(html);
	req->base_uri = g_strdup(base_uri);
	g_idle_add(do_reload, req);
}
