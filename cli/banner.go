package cli

import (
	"blueprint/appinfo"
	"fmt"
	"strings"
)

func center(text string, width int) string {
	padding := width - len(text)
	left := padding / 2
	right := padding - left

	return strings.Repeat(" ", left) +
		text +
		strings.Repeat(" ", right)
}

func left(text string, width int) string {
	padding := width - len(text)
	left := 1
	right := padding - left

	return strings.Repeat(" ", left) +
		text +
		strings.Repeat(" ", right)
}

func PrintBanner() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Printf("║%s║\n", center(appinfo.Name, appinfo.BannerWidth))
	fmt.Printf("║%s║\n", center(appinfo.Version, appinfo.BannerWidth))
	fmt.Printf("║%s║\n", center("", appinfo.BannerWidth))
	fmt.Printf("║%s║\n", left("Commands list = 'blue help'", appinfo.BannerWidth))
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
}
