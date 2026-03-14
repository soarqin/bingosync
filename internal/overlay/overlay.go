// Package overlay provides the embedded OBS overlay HTML page.
package overlay

import _ "embed"

// HTML is the embedded overlay HTML page served to OBS Browser Source clients.
//
//go:embed overlay.html
var HTML []byte
