package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

// EditMode enum type.
type EditMode int

// EditMode enum values.
const (
	EditModeCanvas         EditMode = 0
	EditModeForegroundGrid EditMode = 1
	EditModeBackgroundGrid EditMode = 2
	EditModeInsert         EditMode = 3
)

// Canvas is our main controller holding state.
type Canvas struct {
	Width, Height int
	PosX, PosY    int

	EditMode EditMode

	SettingForeground int
	SettingBackground int

	State [][][3]int
}

// NewCanvas creates the canvas.
func NewCanvas(initialWidth, initialHeight int) *Canvas {
	c := &Canvas{
		Width:  initialWidth,
		Height: initialHeight,

		PosX: 0,
		PosY: 0,

		EditMode: EditModeCanvas,

		SettingForeground: 1,
		SettingBackground: 15,

		State: make([][][3]int, initialHeight),
	}
	for i := range c.State {
		c.State[i] = make([][3]int, initialWidth)
	}
	return c
}

func (c *Canvas) clearScreen() {
	c.PosX = 0
	c.PosY = 0
	rawClearScreen()
}

func (c *Canvas) redraw() {
	c.clearScreen()
	for _, line := range c.State {
		for _, elem := range line {
			if elem[0] == 0 {
				elem[0] = ' '
			}
			fmt.Printf("\033[38;5;%dm\033[48;5;%dm%c\033[0m", elem[1], elem[2], elem[0])
		}
		fmt.Print("\r\n")
	}
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

func rawClearScreen() {
	fmt.Print("\033[2J")
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

func rawEnableMouseClickReporting() {
	fmt.Print("\033[?1000h")
}

func rawDisableMouseClickReporting() {
	fmt.Print("\033[?1000l")
}

func printColorGrid() {
	for i := 0; i < 256; i++ {
		if i > 0 && i%16 == 0 {
			fmt.Print("\r\n")
		}
		fmt.Printf("\033[48;5;%dm  \033[0m", i)
	}
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

// Goals:
// - [x] Canvas with cursor
//   - [x] Display cursor position somewhere
// - [x] Select and set a pixel
// - [x] Set a color for a pixel
// - [x] Allow chars in pixel
// - [x] Save / Load work

func run() error {
	ts, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("makeRaw: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), ts) }() // Best effort.

	rawEnableMouseClickReporting()
	defer rawDisableMouseClickReporting()

	c := NewCanvas(100, 20)
	c.clearScreen()

loop:

	fmt.Print("\0337") // Save cursor position.
	fmt.Printf("\033[%d;%dH", 1, 120-9) // Move cursor.
	// Print UI.
	fmt.Printf("% 9s", fmt.Sprintf("%d/%d", c.PosX, c.PosY))
	fmt.Printf("\033[%d;%dH", 2, 120-6) // Move cursor.
	fmt.Printf("fg: \033[48;5;%dm  \033[0m", c.SettingForeground)
	fmt.Printf("\033[%d;%dH", 3, 120-6) // Move cursor.
	fmt.Printf("bg: \033[48;5;%dm  \033[0m", c.SettingBackground)
	fmt.Printf("\033[%d;%dH", 4, 120-7) // Move cursor.
	fmt.Printf("mode: %d", c.EditMode)
	fmt.Print("\0338") // Restore cursor position.

	buf := make([]byte, 6)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("stdin read: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("unexpected size read")
	}
	buf = buf[:n]

	if n == 6 {
		c.PosX, c.PosY = int(buf[4]-33), int(buf[5]-33)

		fmt.Printf("\033[%d;%dH", c.PosY+1, c.PosX+1) // Set cursor position.
		goto loop
	}

	if c.EditMode == EditModeCanvas {
		if buf[0] == 'q' {
			fmt.Printf("q press detected, exiting.\r\n")
			return nil
		}

		// Check if we hit an arrow key.
		// 0o33: ESC
		// 0o133: '['
		if n == 3 && buf[0] == 0o33 && buf[1] == 0o133 {
			switch buf[2] {
			case 'A': // Up.
				if c.PosY > 0 {
					c.PosY--
				}
			case 'B': // Down.
				c.PosY++ // TODO: Handle upper bounds.
			case 'C': // Right.
				c.PosX++ // TODO: Handle upper bounds.
			case 'D': // Left
				if c.PosX > 0 {
					c.PosX--
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", buf[2])
			}
			fmt.Printf("\033[%c", buf[2])
		}
		if n == 1 && buf[0] == ' ' {
			c.State[c.PosY][c.PosX][0] = ' '
			c.State[c.PosY][c.PosX][1] = c.SettingForeground
			c.State[c.PosY][c.PosX][2] = c.SettingBackground

			fmt.Printf("\033[38;5%dm\033[48;5;%dm \033[0m", c.SettingForeground, c.SettingBackground)
			c.PosX++
		}
		if n == 1 && buf[0] == 'f' {
			c.clearScreen()
			printColorGrid()
			c.EditMode = EditModeForegroundGrid
		}
		if n == 1 && buf[0] == 'b' {
			c.clearScreen()
			printColorGrid()
			c.EditMode = EditModeBackgroundGrid
		}
		if n == 1 && buf[0] == 'i' {
			c.EditMode = EditModeInsert
			goto loop
		}
		if n == 1 && buf[0] == 'o' {
			encodedData, err := os.ReadFile("sprite.sprt")
			if err != nil {
				return fmt.Errorf("readfile: %w", err)
			}

			c.State = nil
			if err := json.Unmarshal(encodedData, &c.State); err != nil {
				return fmt.Errorf("json unmarshal: %w", err)
			}
			c.redraw()
			goto loop
		}
		if n == 1 && buf[0] == 's' {
			encodedData, err := json.Marshal(c.State)
			if err != nil {
				return fmt.Errorf("json marshal: %w", err)
			}

			if err := os.WriteFile("sprite.sprt", encodedData, 0o600); err != nil {
				return fmt.Errorf("writefile: %w", err)
			}
		}
	}

	if c.EditMode != EditModeCanvas {
		if n == 1 && buf[0] == '\033' {
			c.EditMode = EditModeCanvas
			c.redraw()
		}
	}

	if c.EditMode == EditModeInsert {
		if buf[0] != '\033' {
			r, _ := utf8.DecodeRune(buf)
			c.State[c.PosY][c.PosX][0] = int(r)
			c.State[c.PosY][c.PosX][1] = c.SettingForeground
			c.State[c.PosY][c.PosX][2] = c.SettingBackground
			fmt.Printf("\033[38;5;%dm\033[48;5;%dm%c\033[0m", c.SettingForeground, c.SettingBackground, r)
			c.PosX++
		}

		// Check if we hit an arrow key.
		// 0o33: ESC
		// 0o133: '['
		if n == 3 && buf[0] == 0o33 && buf[1] == 0o133 {
			switch buf[2] {
			case 'A': // Up.
				if c.PosY > 0 {
					c.PosY--
				}
			case 'B': // Down.
				c.PosY++ // TODO: Handle upper bounds.
			case 'C': // Right.
				c.PosX++ // TODO: Handle upper bounds.
			case 'D': // Left
				if c.PosX > 0 {
					c.PosX--
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", buf[2])
			}
			fmt.Printf("\033[%c", buf[2])
		}
	}

	if c.EditMode == EditModeForegroundGrid || c.EditMode == EditModeBackgroundGrid {
		if n == 3 && buf[0] == 0o33 && buf[1] == 0o133 {
			switch buf[2] {
			case 'A': // Up.
				if c.PosY > 0 {
					c.PosY--
					fmt.Printf("\033[%c", buf[2])
				}
			case 'B': // Down.
				if c.PosY < 15 {
					c.PosY++ // TODO: Handle upper bounds.
					fmt.Printf("\033[%c", buf[2])
				}
			case 'C': // Right.
				if c.PosX < 30 {
					c.PosX += 2 // TODO: Handle upper bounds.
					fmt.Printf("\033[%c", buf[2])
					fmt.Printf("\033[%c", buf[2])
				}
			case 'D': // Left
				if c.PosX > 0 {
					c.PosX -= 2
					fmt.Printf("\033[%c", buf[2])
					fmt.Printf("\033[%c", buf[2])
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", buf[2])
			}
		}

		if n == 1 && buf[0] == ' ' {
			if c.EditMode == EditModeForegroundGrid {
				c.SettingForeground = c.PosY*16 + c.PosX/2
			} else {
				c.SettingBackground = c.PosY*16 + c.PosX/2
			}
		}
	}

	goto loop
}

func main() {
	if err := run(); err != nil {
		println("Fail:", err.Error())
		return
	}
}
