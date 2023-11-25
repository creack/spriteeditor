package keyboard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"spriteeditor/termseq"

	"golang.org/x/term"
)

// Key enum type.
type Key rune

// Key enum values.
const (
	KeyEscape    Key = 0o33
	KeyBackspace Key = 0o177

	KeyArrowUp    Key = 'A'
	KeyArrowDown  Key = 'B'
	KeyArrowRight Key = 'C'
	KeyArrowLeft  Key = 'D'
)

// Kind enum type.
type Kind int

// Kind enum values.
const (
	kindMin Kind = iota
	KindRegular
	KindSpecial
	KindArrow
	KindMouse
)

// Event represent a mouse/keyboard event.
type Event struct {
	Kind Kind // Mouse/Keyboard/Special key.

	Char Key

	// For mouse events.
	X int
	Y int

	Size int
	Raw  []byte

	Error error
}

// Open the keyboard manager. This sets the terminal in raw mode,
// the caller is expected to call Close() with the returned state
// to restore the terminal modes.
func Open() (*term.State, <-chan Event, error) {
	ts, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, nil, fmt.Errorf("makeRaw: %w", err)
	}
	termseq.EnableMouseClickReporting()

	ch := make(chan Event)
	go readLoop(ch)
	return ts, ch, nil
}

func readLoop(ch chan<- Event) {
	defer close(ch)

	buf := make([]byte, 6)
loop:
	n, err := os.Stdin.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		select {
		case ch <- Event{Error: fmt.Errorf("stdin read: %w", err)}:
		default:
			// Note that if we reach here, we had an error that has been discarded.
		}
		return
	}
	if n == 0 {
		panic(fmt.Errorf("unexpected size read"))
	}

	r, _ := utf8.DecodeRune(buf[:n])

	ev := Event{
		Kind: kindMin,

		Char: Key(r),

		X: 0,
		Y: 0,

		Size: n,
		Raw:  buf[:n],

		Error: nil,
	}

	// Check if we are dealing with a mouse event (CSI + 0o155 + x + y).
	if n == 6 && buf[0] == 0o33 && buf[1] == 0o133 && buf[2] == 0o115 {
		ev.Kind = KindMouse
		ev.X, ev.Y = int(buf[4]-33), int(buf[5]-33)
	}

	// Handle the case of a single regular ESC press.
	if n == 1 && Key(ev.Char) == KeyEscape {
		ev.Kind = KindRegular
	}

	// Check if we have a CSI prefix.
	if n > 2 && buf[0] == byte(KeyEscape) && buf[1] == '[' {
		// CSI mode.
		if n == 3 && buf[2] >= byte(KeyArrowUp) && buf[2] <= byte(KeyArrowLeft) {
			// Directional arrow key pressed.
			ev.Kind = KindArrow
			ev.Char = Key(buf[2])
		}
	}

	if ev.Kind == kindMin {
		ev.Kind = KindRegular
	}

	// Attempt to send the event through the channel.
	// If nothing is ready to receive, discard event.
	select {
	case ch <- ev:
	default:
	}

	goto loop
}

// Close the keyboard maanger, frees resources, restores the terminal.
// NOTE: Also closes os.Stdin, meaning it can't be re-used after Close.
func Close(ts *term.State) error {
	defer func() { _ = os.Stdin.Close() }() // Best effort. Attempts to unblock the read loop.

	termseq.DisableMouseClickReporting()
	return term.Restore(int(os.Stdin.Fd()), ts)
}
