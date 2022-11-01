package tools

import (
	"fmt"
	"runtime"
)

type color int

var (
	Black  color
	White  color
	Red    color
	Blue   color
	Green  color
	Yellow color
	Purple color
)

func init() {
	// darwin, windows, linux
	if runtime.GOOS == "windows" {
		Black = 0
		Blue = 1
		Green = 2
		Red = 4
		Purple = 5
		Yellow = 6
		White = 7
	} else {
		Black = 30
		White = 37
		Red = 31
		Blue = 34
		Green = 32
		Yellow = 33
		Purple = 35
	}
}

func Print(c color, v any) {
	fmt.Print(fmt.Sprintf("%c[%d;%d;%dm%v%c[0m", 0x1B, 0, 0, c, v, 0x1B))
}

func Println(c color, v any) {
	fmt.Println(fmt.Sprintf("%c[%d;%d;%dm%v%c[0m", 0x1B, 0, 0, c, v, 0x1B))
}
