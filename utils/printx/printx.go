package printx

import (
	"fmt"
	"strings"
)

func PrintStandardHeader(header string) {
	hBar := strings.Repeat("-", 80)
	fmt.Println("\n" + hBar + "\n" + header + "\n" + hBar)
}

type Color string

const (
	ColorRed    Color = "red"
	ColorYellow Color = "yellow"
	ColorGreen  Color = "green"
	ColorBlue   Color = "blue"
	ColorPurple Color = "purple"
	ColorCyan   Color = "cyan"
	ColorWhite  Color = "white"
	ColorGray   Color = "gray"
	ColorBlack  Color = "black"
)

func PrintInColor(color Color, text string) {
	switch color {
	case ColorRed:
		fmt.Printf("\033[31m%s\033[0m\n", text)
	case ColorYellow:
		fmt.Printf("\033[33m%s\033[0m\n", text)
	case ColorGreen:
		fmt.Printf("\033[32m%s\033[0m\n", text)
	case ColorBlue:
		fmt.Printf("\033[34m%s\033[0m\n", text)
	case ColorPurple:
		fmt.Printf("\033[35m%s\033[0m\n", text)
	case ColorCyan:
		fmt.Printf("\033[36m%s\033[0m\n", text)
	case ColorWhite:
		fmt.Printf("\033[37m%s\033[0m\n", text)
	case ColorGray:
		fmt.Printf("\033[90m%s\033[0m\n", text)
	case ColorBlack:
		fmt.Printf("\033[30m%s\033[0m\n", text)
	default:
		fmt.Println(text)
	}
}
