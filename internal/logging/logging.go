// Package logging provides a lightweight debug-log toggle shared across the app.
package logging

import "log"

// debugEnabled gates verbose debug output. It is set once at startup via
// SetDebug and read without synchronization thereafter.
var debugEnabled bool

// SetDebug enables or disables debug logging globally.
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

// Debug reports whether debug logging is currently enabled.
func Debug() bool {
	return debugEnabled
}

// Debugf logs a formatted message only when debug logging is enabled.
func Debugf(format string, v ...any) {
	if debugEnabled {
		log.Printf(format, v...)
	}
}

// Debugln logs a message only when debug logging is enabled.
func Debugln(v ...any) {
	if debugEnabled {
		log.Println(v...)
	}
}
