package main

import "fmt"

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
