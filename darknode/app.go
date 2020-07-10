package darknode

import (
	"path/filepath"

	"github.com/fatih/color"
	"github.com/renproject/darknode-cli/darknode/provider"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

// Create new cli application
func App() *cli.App {
	app := cli.NewApp()
	app.Name = "darknode"
	app.Usage = "A command-line tool for managing Ren nodes."
	app.EnableBashCompletion = true

	// Define sub-commands
	app.Commands = []*cli.Command{
		{
			Name:  "up",
			Usage: "Deploy a new Ren node",
			Flags: []cli.Flag{
				// General
				NameFlag, TagsFlag, NetworkFlag, ConfigFlag,
				// AWS
				AwsFlag, AwsAccessKeyFlag, AwsSecretKeyFlag, AwsInstanceFlag, AwsRegionFlag, AwsProfileFlag,
				// Digital Ocean
				DoFlag, DoRegionFlag, DoSizeFlag, DoTokenFlag,
				// Google Cloud Platform
				GcpFlag, GcpRegionFlag, GcpCredFlag, GcpMachineFlag,
			},
			Action: func(ctx *cli.Context) error {
				// Parse the provider and deploy the node
				p, err := provider.ParseProvider(ctx)
				if err != nil {
					return err
				}
				return p.Deploy(ctx)
			},
		},
		{
			Name:    "destroy",
			Usage:   "Destroy one of your Ren node",
			Aliases: []string{"down"},
			Flags:   []cli.Flag{TagsFlag, ForceFlag},
			Action: func(c *cli.Context) error {
				return destroyNode(c)
			},
		},
		{
			Name:  "update",
			Usage: "Update your Ren nodes to the latest version",
			Flags: []cli.Flag{TagsFlag, VersionFlag, DowngradeFlag},
			Action: func(c *cli.Context) error {
				return updateNode(c)
			},
		},
		{
			Name:  "ssh",
			Flags: []cli.Flag{},
			Usage: "SSH into one of your Ren node",
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
				return resize(c)
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
		color.Red("[Warning] run 'darknode --help' for a list of available commands", command)
	}

	return app
}
