package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/renproject/darknode-cli/cmd/provider"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

// This will be populated on build
var binaryVersion = "undefined"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	// Create new cli application
	app := cli.NewApp()
	app.Name = "Nodectl"
	app.Usage = "A command-line tool for managing Ren nodes."
	app.Version = binaryVersion
	app.EnableBashCompletion = true

	// Fetch latest release and check if our version is behind.
	checkUpdates(app.Version)

	// Define sub-commands
	app.Commands = []*cli.Command{
		{
			Name:  "up",
			Usage: "Deploy a new Darknode",
			Flags: []cli.Flag{
				// General
				NameFlag, TagsFlag, NetworkFlag,
				// AWS
				AwsFlag, AwsAccessKeyFlag, AwsSecretKeyFlag, AwsInstanceFlag, AwsRegionFlag, AwsProfileFlag,
				// Digital Ocean
				DoFlag, DoRegionFlag, DoSizeFlag, DoTokenFlag,
				// Google Cloud Platform
				GcpFlag, GcpRegionFlag, GcpCredFlag, GcpMachineFlag,
			},
			Action: func(c *cli.Context) error {
				p, err := provider.ParseProvider(c)
				if err != nil {
					return err
				}
				return p.Deploy(c)
			},
		},
		{
			Name:    "destroy",
			Usage:   "Destroy one of your Darknode",
			Aliases: []string{"down"},
			Flags:   []cli.Flag{TagsFlag, ForceFlag},
			Action: func(c *cli.Context) error {
				return destroyNode(c)
			},
		},
		{
			Name:  "update",
			Usage: "Update your Darknodes to the latest software and configuration",
			Flags: []cli.Flag{TagsFlag, VersionFlag, DowngradeFlag},
			Action: func(c *cli.Context) error {
				return updateNode(c)
			},
		},
		{
			Name:  "ssh",
			Flags: []cli.Flag{},
			Usage: "SSH into one of your Darknode",
			Action: func(c *cli.Context) error {
				name := c.Args().First()
				if err := util.ValidateNodeName(name); err != nil {
					return err
				}
				ip, err := util.IP(name)
				if err != nil {
					return err
				}
				keyPath := filepath.Join(util.NodePath(name), "ssh_keypair")
				return util.Run("ssh", "-i", keyPath, "darknode@"+ip, "-oStrictHostKeyChecking=no")
			},
		},
		{
			Name:  "start",
			Flags: []cli.Flag{TagsFlag},
			Usage: "Start a single Darknode or a set of Darknodes by its tag",
			Action: func(c *cli.Context) error {
				return updateServiceStatus(c, "start")
			},
		},
		{
			Name:  "stop",
			Flags: []cli.Flag{TagsFlag},
			Usage: "Stop a single Darknode or a set of Darknodes by its tag",
			Action: func(c *cli.Context) error {
				return updateServiceStatus(c, "stop")
			},
		},
		{
			Name:  "restart",
			Flags: []cli.Flag{TagsFlag},
			Usage: "Restart a single Darknode or a set of Darknodes by its tag",
			Action: func(c *cli.Context) error {
				return updateServiceStatus(c, "restart")
			},
		},
		{
			Name:  "list",
			Usage: "List information about all of your Darknodes",
			Flags: []cli.Flag{TagsFlag},
			Action: func(c *cli.Context) error {
				return listAllNodes(c)
			},
		},
		{
			Name:  "withdraw",
			Usage: "Withdraw all the ETH and REN the Darknode address holds",
			Flags: []cli.Flag{AddressFlag},
			Action: func(c *cli.Context) error {
				return withdraw(c)
			},
		},
		{
			Name:  "resize",
			Usage: "Resize the instance type of a specific darknode",
			Flags: []cli.Flag{InstanceFlag, StorageFlag},
			Action: func(c *cli.Context) error {
				return expand(c)
			},
		},
		{
			Name:  "exec",
			Usage: "Execute script on Darknodes",
			Flags: []cli.Flag{TagsFlag, ScriptFlag, FileFlag},
			Action: func(c *cli.Context) error {
				return execScript(c)
			},
		},
		{
			Name:  "register",
			Usage: "Redirect you to the register page of a particular darknode",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				name := c.Args().First()
				if err := util.ValidateNodeName(name); err != nil {
					return err
				}

				url, err := util.RegisterUrl(name)
				if err != nil {
					return err
				}
				color.Green("If the browser doesn't open for you, please copy the following url and open in browser.")
				color.Green(url)
				return util.OpenInBrowser(url)
			},
		},
	}

	// Show error message and display the help page when command is not found.
	app.CommandNotFound = func(c *cli.Context, command string) {
		color.Red("[Warning] command '%q' not found", command)
		color.Red("[Warning] run 'nodectl --help' for a list of available commands", command)
	}

	// Start the app
	err := app.Run(os.Args)
	if err != nil {
		// Remove the timestamp for error message
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
		color.Red(err.Error())
		os.Exit(1)
	}
}

// checkUpdates fetches the latest release of `nodectl` from github and compares the versions. It warns the user if
// current version is out of date.
func checkUpdates(curVer string) {
	// Get latest release
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
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
		color.Red("You can update your nodectl with `nodectl upgrade` command")
	}
}
