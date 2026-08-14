package cli

import (
	"blueprint/appinfo"
	"blueprint/models"
	"flag"
	"fmt"
	"os"
)

var (
	versionShort = flag.Bool("v", false, "Show version")
	versionLong  = flag.Bool("version", false, "Show version")
	helpShort    = flag.Bool("h", false, "List of all commands")
	helpLong     = flag.Bool("help", false, "List of all commands")
)

func Flags() {
	flag.Parse()

	version()
	help()

}

func version() {
	if *versionShort || *versionLong {
		fmt.Printf("%s %s\n", appinfo.Name, appinfo.Version)
		os.Exit(0)
	}
}

func Version(input models.CommandInput) {
	fmt.Printf("%s %s\n", appinfo.Name, appinfo.Version)
}

func help() {
	if *helpShort || *helpLong {
		Help()
		os.Exit(0)
	}
}
