package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"spriteeditor/keyboard"
)

func printCommonChars() {
	buf, err := os.ReadFile("commonchars.txt")
	if err != nil {
		panic(err)
	}
	content := string(buf)

	for i, line := range strings.Split(content, "\n") {
		if len(line) == 0 {
			continue
		}
		if i > 0 && i%32 == 0 {
			fmt.Printf("\r\n\r\n")
		}
		r, _ := utf8.DecodeRuneInString(line)
		fmt.Printf("%c ", r)
	}
}

func (c *Canvas) modeCommonChars(ev keyboard.Event) error {
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
