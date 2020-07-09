package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/renproject/darknode-cli/darknode"
	"github.com/renproject/darknode-cli/util"
)

// This will be populated on build
var binaryVersion = "undefined"

// Randomize the seed.
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	app := darknode.App()
	app.Version = binaryVersion

	// Fetch latest release and check if our version is behind.
	checkUpdates(binaryVersion)

	// Start the app
	err := app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

// checkUpdates fetches the latest release of `darknode-cli` from github and compares the versions. It warns the user if
// current version is out of date.
func checkUpdates(curVer string) {
	// Get latest release
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Compare versions
	versionLatest, err := util.CurrentReleaseVersion(ctx)
	if err != nil {
		return
	}
	versionCurrent, err := version.NewVersion(curVer)
	if err != nil {
		color.Red("cannot parse current software version, err = %v", err)
		return
	}

	// Warn user they're using a older version.
	if versionCurrent.LessThan(versionLatest) {
		color.Red("You are running %v", curVer)
		color.Red("A new release is available (%v)", versionLatest.String())
		color.Red("You can update your darknode-cli with `darknode self update` command")
	}
}
