package main

import (
	"fmt"

	"spriteeditor/termseq"
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
	termseq.ClearScreen()
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

func (c *Canvas) printUI() {
	termseq.SaveCursor()

	termseq.MoveCursor(120-9, 1)
	fmt.Printf("% 9s", fmt.Sprintf("%d/%d", c.PosX, c.PosY))

	termseq.MoveCursor(120-6, 2)
	fmt.Printf("fg: %s", termseq.WrapColor(0, c.SettingForeground, "  "))

	termseq.MoveCursor(120-6, 3)
	fmt.Printf("bg: %s", termseq.WrapColor(0, c.SettingBackground, "  "))

	termseq.MoveCursor(120-7, 4)
	fmt.Printf("mode: %d", c.EditMode)

	termseq.RestoreCursor()
}
