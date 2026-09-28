// This file is the cgo bridge to the GTK 4 / WebKitGTK window implemented in
// viewer.c. All GTK calls happen on the thread that runs RunViewer; other
// goroutines may only call ReloadViewer, which hands work over to the GTK main
// loop.
package main

/*
#cgo pkg-config: gtk4 webkitgtk-6.0
#include <stdlib.h>
#include "viewer.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// RunViewer opens the window showing html and blocks until it is closed.
// baseURI is used to resolve relative resource paths such as images.
func RunViewer(title, html, baseURI string) error {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	cURI := C.CString(baseURI)
	defer C.free(unsafe.Pointer(cURI))

	if C.viewer_run(cTitle, cHTML, cURI) != 0 {
		return errors.New("cannot initialise GTK (no display?)")
	}
	return nil
}

// ReloadViewer replaces the displayed HTML and keeps the scroll position.
// Safe to call from any goroutine.
func ReloadViewer(html, baseURI string) {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	cURI := C.CString(baseURI)
	defer C.free(unsafe.Pointer(cURI))

	C.viewer_reload(cHTML, cURI)
}
