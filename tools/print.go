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
	// darwin, linux
	Black = 30
	White = 37
	Red = 31
	Blue = 34
	Green = 32
	Yellow = 33
	Purple = 35
}

func Print(c color, v any) {
	if runtime.GOOS == "windows" {
		fmt.Print(v)
	} else {
		fmt.Print(fmt.Sprintf("%c[%d;%d;%dm%v%c[0m", 0x1B, 0, 0, c, v, 0x1B))
	}
}

func Println(c color, v any) {
	if runtime.GOOS == "windows" {
		fmt.Println(v)
	} else {
		fmt.Println(fmt.Sprintf("%c[%d;%d;%dm%v%c[0m", 0x1B, 0, 0, c, v, 0x1B))
	}
}
