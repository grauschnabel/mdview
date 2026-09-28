package main

/*
#cgo pkg-config: webkit2gtk-4.1 gtk+-3.0
#include <webkit2/webkit2.h>
#include <gtk/gtk.h>
#include <stdlib.h>
#include <stdint.h>

// Scroll position to restore once the next load has finished.
static double pending_scroll = 0;

// on_load_changed restores the scroll position after a reload.
// Uses WebKit's own script API, which is not subject to the document's CSP.
static void on_load_changed(WebKitWebView *wv, WebKitLoadEvent event, gpointer user_data) {
	if (event == WEBKIT_LOAD_FINISHED && pending_scroll > 0) {
		char js[64];
		g_snprintf(js, sizeof js, "window.scrollTo(0, %d);", (int)pending_scroll);
		webkit_web_view_evaluate_javascript(wv, js, -1, NULL, NULL, NULL, NULL, NULL);
		pending_scroll = 0;
	}
}

// new_webview creates a WebKitWebView with hardened settings:
// external navigation blocked. JavaScript is enabled only so the app itself
// can read/restore the scroll position; scripts inside the rendered document
// never run because the page CSP is default-src 'none'.
static GtkWidget* new_webview() {
	WebKitSettings *settings = webkit_settings_new();
	webkit_settings_set_enable_javascript(settings, TRUE);

	WebKitWebView *wv = WEBKIT_WEB_VIEW(webkit_web_view_new_with_settings(settings));
	g_signal_connect(wv, "load-changed", G_CALLBACK(on_load_changed), NULL);

	return GTK_WIDGET(wv);
}

typedef struct {
	GtkWidget *widget;
	char *html;
	char *base_uri;
} reload_req;

static void on_scroll_read(GObject *source, GAsyncResult *res, gpointer user_data) {
	reload_req *req = user_data;
	JSCValue *r = webkit_web_view_evaluate_javascript_finish(WEBKIT_WEB_VIEW(source), res, NULL);
	if (r) {
		pending_scroll = jsc_value_to_double(r);
		g_object_unref(r);
	}
	webkit_web_view_load_html(WEBKIT_WEB_VIEW(req->widget), req->html, req->base_uri);
	g_free(req->html);
	g_free(req->base_uri);
	g_free(req);
}

// webview_reload loads new HTML and keeps the current scroll position.
// Must be called on the GTK main thread.
static void webview_reload(GtkWidget *widget, const char *html, const char *base_uri) {
	reload_req *req = g_new0(reload_req, 1);
	req->widget = widget;
	req->html = g_strdup(html);
	req->base_uri = g_strdup(base_uri);
	webkit_web_view_evaluate_javascript(WEBKIT_WEB_VIEW(widget), "window.scrollY;", -1, NULL, NULL, NULL, on_scroll_read, req);
}

// webview_load_html loads HTML content into the WebView with the given base URI.
static void webview_load_html(GtkWidget *widget, const char *html, const char *base_uri) {
	webkit_web_view_load_html(WEBKIT_WEB_VIEW(widget), html, base_uri);
}

// block_navigation is the callback for the "decide-policy" signal.
// It blocks all navigation except the initial HTML load.
static gboolean block_navigation(WebKitWebView *wv,
                                  WebKitPolicyDecision *decision,
                                  WebKitPolicyDecisionType type,
                                  gpointer user_data) {
	if (type == WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION) {
		WebKitNavigationPolicyDecision *nav = WEBKIT_NAVIGATION_POLICY_DECISION(decision);
		WebKitNavigationAction *action = webkit_navigation_policy_decision_get_navigation_action(nav);
		WebKitNavigationType nav_type = webkit_navigation_action_get_navigation_type(action);

		// Allow only "other" navigation type (initial load).
		if (nav_type != WEBKIT_NAVIGATION_TYPE_OTHER) {
			webkit_policy_decision_ignore(decision);
			return TRUE;
		}
	}
	return FALSE;
}

// connect_block_navigation connects the decide-policy signal to block external navigation.
static void connect_block_navigation(GtkWidget *widget) {
	g_signal_connect(WEBKIT_WEB_VIEW(widget), "decide-policy",
	                 G_CALLBACK(block_navigation), NULL);
}

// add_webview_to_window adds the WebView widget to a GTK container (window).
// The window arrives as a plain address (gotk3's Native()), converted here in C.
static void add_webview_to_window(uintptr_t window, GtkWidget *webview) {
	gtk_container_add(GTK_CONTAINER((GtkWidget *)window), webview);
}
*/
import "C"
import "unsafe"

// webviewWidget holds the C pointer to the WebKitWebView GtkWidget.
// GTK manages the widget lifecycle; Go only holds a reference.
type webviewWidget struct {
	ptr *C.GtkWidget
}

// NewWebView creates a new WebKitGTK WebView widget with hardened security settings.
func NewWebView() *webviewWidget {
	w := C.new_webview()
	C.connect_block_navigation(w)
	return &webviewWidget{ptr: w}
}

// LoadHTML loads rendered HTML into the WebView with the given base URI
// for resolving relative resource paths.
func (wv *webviewWidget) LoadHTML(html, baseURI string) {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	cURI := C.CString(baseURI)
	defer C.free(unsafe.Pointer(cURI))

	C.webview_load_html(wv.ptr, cHTML, cURI)
}

// Reload replaces the displayed HTML and restores the scroll position.
// Must be called on the GTK main thread.
func (wv *webviewWidget) Reload(html, baseURI string) {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	cURI := C.CString(baseURI)
	defer C.free(unsafe.Pointer(cURI))

	C.webview_reload(wv.ptr, cHTML, cURI)
}

// AddToWindow adds the WebView to a GTK window container.
func (wv *webviewWidget) AddToWindow(window uintptr) {
	C.add_webview_to_window(C.uintptr_t(window), wv.ptr)
}
