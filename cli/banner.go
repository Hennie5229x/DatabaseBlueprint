package cli

import (
	"blueprint/appinfo"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	Reset = "\033[0m"
	Cyan  = "\033[96m"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func cyan(text string) string {
	return Cyan + text + Reset
}

func textWidth(text string) int {
	plain := ansiRegex.ReplaceAllString(text, "")
	return utf8.RuneCountInString(plain)
}

func center(text string, width int) string {
	padding := width - textWidth(text)

	if padding < 0 {
		padding = 0
	}

	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) +
		text +
		strings.Repeat(" ", rightPadding)
}

func left(text string, width int) string {
	padding := width - textWidth(text)

	if padding < 1 {
		padding = 1
	}

	leftPadding := 1
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) +
		text +
		strings.Repeat(" ", rightPadding)
}

func PrintBanner() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Printf("║%s║\n", center(cyan("◆ "+appinfo.Name+" ◆"), appinfo.BannerWidth))
	fmt.Printf("║%s║\n", center(appinfo.Version, appinfo.BannerWidth))
	fmt.Printf("║%s║\n", center("", appinfo.BannerWidth))
	fmt.Printf("║%s║\n", left("➜ blue help", appinfo.BannerWidth))
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
}
