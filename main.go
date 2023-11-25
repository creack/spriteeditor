package main

import (
	"encoding/json"
	"fmt"
	"os"

	"spriteeditor/keyboard"
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
	ts, ch, err := keyboard.Open()
	if err != nil {
		return fmt.Errorf("keyboard open: %w", err)
	}
	defer func() { _ = keyboard.Close(ts) }() // Best effort.

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

	ev := <-ch

	// if ev.Char == 'q' {
	// 	return nil
	// }
	// fmt.Printf("[%d] %O\r\n", ev.Size, ev.Raw)

	// goto loop

	if ev.Kind == keyboard.KindMouse {
		c.PosX, c.PosY = ev.X, ev.Y

		fmt.Printf("\033[%d;%dH", c.PosY+1, c.PosX+1) // Set cursor position.
		goto loop
	}

	if c.EditMode == EditModeCanvas {
		// Check if we hit an arrow key.
		if ev.Kind == keyboard.KindArrow {
			switch ev.Char {
			case keyboard.KeyArrowUp: // Up.
				if c.PosY > 0 {
					c.PosY--
				}
			case keyboard.KeyArrowDown: // Down.
				c.PosY++ // TODO: Handle upper bounds.
			case keyboard.KeyArrowRight: // Right.
				c.PosX++ // TODO: Handle upper bounds.
			case keyboard.KeyArrowLeft: // Left
				if c.PosX > 0 {
					c.PosX--
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", ev.Char)
			}
			fmt.Printf("\033[%c", ev.Char)
		}

		if ev.Kind == keyboard.KindRegular {
			switch ev.Char {
			case 'q':
				fmt.Printf("q press detected, exiting.\r\n")
				return nil

			case ' ':
				c.State[c.PosY][c.PosX][0] = ' '
				c.State[c.PosY][c.PosX][1] = c.SettingForeground
				c.State[c.PosY][c.PosX][2] = c.SettingBackground

				fmt.Printf("\033[38;5%dm\033[48;5;%dm \033[0m", c.SettingForeground, c.SettingBackground)
				c.PosX++

			case 'f':
				c.clearScreen()
				printColorGrid()
				c.EditMode = EditModeForegroundGrid

			case 'b':
				c.clearScreen()
				printColorGrid()
				c.EditMode = EditModeBackgroundGrid

			case 'i':
				c.EditMode = EditModeInsert
				goto loop

			case 'o':
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

			case 's':
				encodedData, err := json.Marshal(c.State)
				if err != nil {
					return fmt.Errorf("json marshal: %w", err)
				}

				if err := os.WriteFile("sprite.sprt", encodedData, 0o600); err != nil {
					return fmt.Errorf("writefile: %w", err)
				}
			}
		}
	}

	if c.EditMode != EditModeCanvas {
		if ev.Kind == keyboard.KindRegular && ev.Char == keyboard.KeyEscape {
			c.EditMode = EditModeCanvas
			c.redraw()
		}
	}

	if c.EditMode == EditModeInsert {
		if ev.Kind == keyboard.KindRegular && ev.Char == keyboard.KeyEscape {
			c.State[c.PosY][c.PosX][0] = int(ev.Char)
			c.State[c.PosY][c.PosX][1] = c.SettingForeground
			c.State[c.PosY][c.PosX][2] = c.SettingBackground
			fmt.Printf("\033[38;5;%dm\033[48;5;%dm%c\033[0m", c.SettingForeground, c.SettingBackground, ev.Char)
			c.PosX++
		}

		// Check if we hit an arrow key.
		if ev.Kind == keyboard.KindArrow {
			switch ev.Char {
			case keyboard.KeyArrowUp: // Up.
				if c.PosY > 0 {
					c.PosY--
				}
			case keyboard.KeyArrowDown: // Down.
				c.PosY++ // TODO: Handle upper bounds.
			case keyboard.KeyArrowRight: // Right.
				c.PosX++ // TODO: Handle upper bounds.
			case keyboard.KeyArrowLeft: // Left
				if c.PosX > 0 {
					c.PosX--
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", ev.Char)
			}
			fmt.Printf("\033[%c", ev.Char)
		}
	}

	if c.EditMode == EditModeForegroundGrid || c.EditMode == EditModeBackgroundGrid {
		// Check if we hit an arrow key.
		if ev.Kind == keyboard.KindArrow {
			switch ev.Char {
			case keyboard.KeyArrowUp: // Up.
				if c.PosY > 0 {
					c.PosY--
					fmt.Printf("\033[%c", ev.Char)
				}
			case keyboard.KeyArrowDown: // Down.
				if c.PosY < 15 {
					c.PosY++ // TODO: Handle upper bounds.
					fmt.Printf("\033[%c", ev.Char)
				}
			case keyboard.KeyArrowRight: // Right.
				if c.PosX < 30 {
					c.PosX += 2 // TODO: Handle upper bounds.
					fmt.Printf("\033[%c", ev.Char)
					fmt.Printf("\033[%c", ev.Char)
				}
			case keyboard.KeyArrowLeft: // Left
				if c.PosX > 0 {
					c.PosX -= 2
					fmt.Printf("\033[%c", ev.Char)
					fmt.Printf("\033[%c", ev.Char)
				}
			default:
				return fmt.Errorf("unexpected arrow key value: %O", ev.Char)
			}
		}

		if ev.Kind == keyboard.KindRegular && ev.Char == ' ' {
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
