#include <gtk/gtk.h>
#include <webkit/webkit.h>

#include "viewer.h"

static GMainLoop *loop;
static WebKitWebView *webview;

// Scroll position to restore once the next load has finished.
static double pending_scroll = 0;

// on_load_changed restores the scroll position after a reload.
static void on_load_changed(WebKitWebView *wv, WebKitLoadEvent event, gpointer user_data) {
	if (event == WEBKIT_LOAD_FINISHED && pending_scroll > 0) {
		char js[64];
		g_snprintf(js, sizeof js, "window.scrollTo(0, %d);", (int)pending_scroll);
		webkit_web_view_evaluate_javascript(wv, js, -1, NULL, NULL, NULL, NULL, NULL);
		pending_scroll = 0;
	}
}

// block_navigation blocks everything except the programmatic load.
static gboolean block_navigation(WebKitWebView *wv, WebKitPolicyDecision *decision,
                                 WebKitPolicyDecisionType type, gpointer user_data) {
	if (type == WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION) {
		WebKitNavigationPolicyDecision *nav = WEBKIT_NAVIGATION_POLICY_DECISION(decision);
		WebKitNavigationAction *action = webkit_navigation_policy_decision_get_navigation_action(nav);

		if (webkit_navigation_action_get_navigation_type(action) != WEBKIT_NAVIGATION_TYPE_OTHER) {
			webkit_policy_decision_ignore(decision);
			return TRUE;
		}
	}
	return FALSE;
}

// on_key_pressed: q, Ctrl+Q and Ctrl+W quit.
static gboolean on_key_pressed(GtkEventControllerKey *ctrl, guint keyval, guint keycode,
                               GdkModifierType state, gpointer user_data) {
	state &= gtk_accelerator_get_default_mod_mask();
	if ((keyval == GDK_KEY_q && (state == 0 || state == GDK_CONTROL_MASK)) ||
	    (keyval == GDK_KEY_w && state == GDK_CONTROL_MASK)) {
		g_main_loop_quit(loop);
		return TRUE;
	}
	return FALSE;
}

static gboolean on_close_request(GtkWindow *win, gpointer user_data) {
	g_main_loop_quit(loop);
	return FALSE;
}

int viewer_run(const char *title, const char *html, const char *base_uri) {
	if (!gtk_init_check()) {
		return 1;
	}

	GtkWidget *window = gtk_window_new();
	gtk_window_set_title(GTK_WINDOW(window), title);
	gtk_window_set_default_size(GTK_WINDOW(window), 900, 700);
	g_signal_connect(window, "close-request", G_CALLBACK(on_close_request), NULL);

	// Hardened settings: external navigation is blocked. JavaScript is enabled
	// only so the app itself can read/restore the scroll position; scripts in the
	// rendered document never run because the page CSP is default-src 'none'.
	webview = WEBKIT_WEB_VIEW(webkit_web_view_new());
	WebKitSettings *settings = webkit_web_view_get_settings(webview);
	webkit_settings_set_enable_javascript(settings, TRUE);
	g_signal_connect(webview, "decide-policy", G_CALLBACK(block_navigation), NULL);
	g_signal_connect(webview, "load-changed", G_CALLBACK(on_load_changed), NULL);
	gtk_window_set_child(GTK_WINDOW(window), GTK_WIDGET(webview));

	GtkEventController *keys = gtk_event_controller_key_new();
	gtk_event_controller_set_propagation_phase(keys, GTK_PHASE_CAPTURE);
	g_signal_connect(keys, "key-pressed", G_CALLBACK(on_key_pressed), NULL);
	gtk_widget_add_controller(window, keys);

	webkit_web_view_load_html(webview, html, base_uri);
	gtk_window_present(GTK_WINDOW(window));

	loop = g_main_loop_new(NULL, FALSE);
	g_main_loop_run(loop);
	g_main_loop_unref(loop);
	return 0;
}

typedef struct {
	char *html;
	char *base_uri;
} reload_req;

static void free_req(reload_req *req) {
	g_free(req->html);
	g_free(req->base_uri);
	g_free(req);
}

static void on_scroll_read(GObject *source, GAsyncResult *res, gpointer user_data) {
	reload_req *req = user_data;
	JSCValue *r = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(source), res, NULL);
	if (r) {
		pending_scroll = jsc_value_to_double(r);
		g_object_unref(r);
	}
	webkit_web_view_load_html(WEBKIT_WEB_VIEW(source), req->html, req->base_uri);
	free_req(req);
}

// do_reload runs on the GTK main thread.
static gboolean do_reload(gpointer data) {
	reload_req *req = data;
	if (!webview) {
		free_req(req);
		return G_SOURCE_REMOVE;
	}
	webkit_web_view_evaluate_javascript(webview, "window.scrollY;", -1, NULL, NULL, NULL,
	                                    on_scroll_read, req);
	return G_SOURCE_REMOVE;
}

void viewer_reload(const char *html, const char *base_uri) {
	reload_req *req = g_new0(reload_req, 1);
	req->html = g_strdup(html);
	req->base_uri = g_strdup(base_uri);
	g_idle_add(do_reload, req);
}
