package ui

import (
	"fmt"
	"time"
)

// Status represents the health state of a project
type Status int

const (
	StatusFresh  Status = iota // green
	StatusWarning              // yellow
	StatusStale                // red
)

// GetStatus returns the status for a given last touched time
func GetStatus(lastTouched time.Time) Status {
	days := int(time.Since(lastTouched).Hours() / 24)
	if days >= 10 {
		return StatusStale
	}
	if days >= 5 {
		return StatusWarning
	}
	return StatusFresh
}

// StatusDot returns a colored Unicode status indicator
func StatusDot(lastTouched time.Time) string {
	status := GetStatus(lastTouched)

	switch status {
	case StatusFresh:
		return "\033[32m🟢\033[0m"
	case StatusWarning:
		return "\033[33m🟡\033[0m"
	case StatusStale:
		return "\033[31m🔴\033[0m"
	default:
		return "⚪"
	}
}

// StatusDotPlain returns a simple text version (no color/unicode)
func StatusDotPlain(lastTouched time.Time) string {
	status := GetStatus(lastTouched)
	switch status {
	case StatusFresh:
		return "[OK]"
	case StatusWarning:
		return "[!]"
	case StatusStale:
		return "[X]"
	default:
		return "[?]"
	}
}

// ColorText wraps text with ANSI color
func ColorText(text string, colorCode int) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", colorCode, text)
}

// Header prints a nice cyber-style header
func Header(title string) {
	fmt.Printf("\033[36m═══ %s ═══\033[0m\n", title)
}

// Divider prints a subtle line
func Divider() {
	fmt.Println("\033[36m────────────────────────────────────────\033[0m")
}