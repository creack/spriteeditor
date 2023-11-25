package keyboard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

// Event represent a mouse/keyboard event.
type Event struct {
	Kind int // Mouse/Keyboard/Special key.

	Char rune

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
	rawEnableMouseClickReporting()

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
		Kind: 0,

		Char: r,

		X: 0,
		Y: 0,

		Size: n,
		Raw:  buf[:n],

		Error: nil,
	}

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
	rawDisableMouseClickReporting()
	return term.Restore(int(os.Stdin.Fd()), ts)
}

func rawEnableMouseClickReporting() {
	fmt.Print("\033[?1000h")
}

func rawDisableMouseClickReporting() {
	fmt.Print("\033[?1000l")
}
