package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"spriteeditor/keyboard"
)

func (c *Canvas) modeCanvas(ev keyboard.Event) error {
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
			return io.EOF

		case ' ':
			c.State[c.PosY][c.PosX][0] = ' '
			c.State[c.PosY][c.PosX][1] = c.SettingForeground
			c.State[c.PosY][c.PosX][2] = c.SettingBackground

			fmt.Printf("\033[38;5%dm\033[48;5;%dm \033[0m", c.SettingForeground, c.SettingBackground)
			c.PosX++

		case 'c':
			c.EditMode = EditModeCommonChars
			c.clearScreen()
			printCommonChars()

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
			return nil

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
			return nil

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

	return nil
}
