package cli

import (
	"os"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// colorEnabled returns whether color output is enabled.
func colorEnabled() bool {
	// --no-color flag
	if flagNoColor {
		return false
	}
	// NO_COLOR environment variable (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// TTY detection
	return isTerminal()
}

// colorize applies an ANSI color to text.
// Returns the original text if color is disabled.
func colorize(color, text string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + colorReset
}

// Convenience functions
func redText(text string) string    { return colorize(colorRed, text) }
func greenText(text string) string  { return colorize(colorGreen, text) }
func yellowText(text string) string { return colorize(colorYellow, text) }
func blueText(text string) string   { return colorize(colorBlue, text) }
func boldText(text string) string   { return colorize(colorBold, text) }
func dimText(text string) string    { return colorize(colorDim, text) }
