package main

import (
	"fmt"

	"spriteeditor/keyboard"
)

func printColorGrid() {
	for i := 0; i < 256; i++ {
		if i > 0 && i%16 == 0 {
			fmt.Print("\r\n")
		}
		fmt.Printf("\033[48;5;%dm  \033[0m", i)
	}
	fmt.Printf("\033[%d;%dH", 0, 0) // Reset cursor.
}

func (c *Canvas) modeColorGrid(ev keyboard.Event, fg bool) error {
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
		if fg {
			c.SettingForeground = c.PosY*16 + c.PosX/2
		} else {
			c.SettingBackground = c.PosY*16 + c.PosX/2
		}
	}
	return nil
}
