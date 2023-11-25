package main

import (
	"errors"
	"fmt"
	"io"

	"spriteeditor/keyboard"
	"spriteeditor/termseq"
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

func run() error {
	ts, ch, err := keyboard.Open()
	if err != nil {
		return fmt.Errorf("keyboard open: %w", err)
	}
	defer func() { _ = keyboard.Close(ts) }() // Best effort.

	c := NewCanvas(100, 20)
	c.clearScreen()
loop:

	c.printUI()

	ev := <-ch

	if ev.Kind == keyboard.KindMouse {
		c.PosX, c.PosY = ev.X, ev.Y

		// Our canvas is 0 indexed but the terminal is 1 indexed.
		termseq.MoveCursor(c.PosX+1, c.PosY+1)

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

	if c.EditMode != EditModeCanvas {
		if ev.Kind == keyboard.KindRegular && ev.Char == keyboard.KeyEscape {
			c.EditMode = EditModeCanvas
			c.redraw()
		}
		goto loop
	}

	goto loop
}

func main() {
	if err := run(); err != nil {
		println("Fail:", err.Error())
		return
	}
}
