package main

import (
	"errors"
	"fmt"
	"io"

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

func rawClearScreen() {
	fmt.Print("\033[2J")
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

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

	if ev.Kind == keyboard.KindMouse {
		c.PosX, c.PosY = ev.X, ev.Y

		fmt.Printf("\033[%d;%dH", c.PosY+1, c.PosX+1) // Set cursor position.
		goto loop
	}
	if c.EditMode != EditModeCanvas {
		if ev.Kind == keyboard.KindRegular && ev.Char == keyboard.KeyEscape {
			c.EditMode = EditModeCanvas
			c.redraw()
		}
		goto loop
	}

	switch c.EditMode {
	case EditModeCanvas:
		err = c.modeCanvas(ev)
	case EditModeInsert:
		err = c.modeInsert(ev)
	case EditModeForegroundGrid, EditModeBackgroundGrid:
		err = c.modeColorGrid(ev, c.EditMode == EditModeForegroundGrid)
	}

	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("handle mode %d: %w", c.EditMode, err)
	}

	goto loop
}

func main() {
	if err := run(); err != nil {
		println("Fail:", err.Error())
		return
	}
}
