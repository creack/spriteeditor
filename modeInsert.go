package main

import (
	"fmt"

	"spriteeditor/keyboard"
)

func (c *Canvas) modeInsert(ev keyboard.Event) error {
	if ev.Kind == keyboard.KindRegular {
		c.State[c.PosY][c.PosX][0] = int(ev.Char)
		c.State[c.PosY][c.PosX][1] = c.SettingForeground
		c.State[c.PosY][c.PosX][2] = c.SettingBackground
		fmt.Printf("\033[38;5;%dm\033[48;5;%dm%c\033[0m", c.SettingForeground, c.SettingBackground, ev.Char)
		c.PosX++
		return nil
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
	return nil
}
