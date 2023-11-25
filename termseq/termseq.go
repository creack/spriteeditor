package termseq

import "fmt"

// ClearScreen clears the screen and resets the cursor to 0/0 origin.
func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

// EnableMouseClickReporting enables mouse reporting so stdin read can get mouse data.
func EnableMouseClickReporting() {
	fmt.Print("\033[?1000h")
}

// DisableMouseClickReporting disables the mosue reporting.
func DisableMouseClickReporting() {
	fmt.Print("\033[?1000l")
}

// SaveCursor saves the current cursor position in the terminal memory.
func SaveCursor() {
	fmt.Print("\0337") // Save cursor position.
}

// RestoreCursor restores the saved cursor position.
func RestoreCursor() {
	fmt.Print("\0338") // Restore cursor position.
}

// MoveCursor to the given coordinates.
func MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x) // Move cursor.
}

// StartFgColor returns the ansi sequence to start a foreground color (256 colors mode).
func StartFgColor(n int) string {
	return fmt.Sprintf("\033[38;5;%dm", n)
}

// StartFgColor returns the ansi sequence to start a background color (256 colors mode).
func StartBgColor(n int) string {
	return fmt.Sprintf("\033[48;5;%dm", n)
}

func ResetColor() string {
	return "\033[0m"
}

// WrapColor wraps the given text with the given fg/bg color and reset ansi sequence.
func WrapColor(fg, bg int, text string) string {
	return StartFgColor(fg) + StartBgColor(bg) + text + ResetColor()
}
