package main

import (
	"blueprint/cli"
)

func main() {
	cli.Flags()   // Handle flags
	cli.Execute() // Handle input Commands
}
