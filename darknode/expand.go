package darknode

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/fatih/color"
	"github.com/renproject/darknode-cli/provider"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

func expand(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if err := util.ValidateNodeName(name); err != nil {
		return err
	}
	storageArg := ctx.Args().Get(1)
	storage, err := strconv.Atoi(storageArg)
	if err != nil {
		return err
	}

	// Fetch the cloud provider
	p, err := provider.GetProvider(name)
	if err != nil {
		return err
	}

	var change Change
	switch p {
	case provider.NameAws:
		change = Change{`volume_size = \d+`, fmt.Sprintf("volume_size = %v", storage)}
	case provider.NameDo:
		// todo
	case provider.NameGcp:
		// todo
	default:
		panic("unknown provider")
	}

	// Read the config file
	path := filepath.Join(util.NodePath(name), "main.tf")
	tf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	original := make([]byte, len(tf))
	copy(original, tf)

	// Update the config file with the changes we want to make
	reg, err := regexp.Compile(change.Regex)
	if err != nil {
		return err
	}

	tf = reg.ReplaceAll(tf, []byte(change.Replacement))
	if err := ioutil.WriteFile(path, tf, 0600); err != nil {
		return err
	}

	// Apply the changes
	color.Green("Resizing dark nodes ...")
	apply := fmt.Sprintf("cd %v && terraform apply -auto-approve -no-color", util.NodePath(name))
	if err := util.Run("bash", "-c", apply); err != nil {
		// revert the `main.tf` file if fail to resize the droplet
		if err := ioutil.WriteFile(path, original, 0600); err != nil {
			return fmt.Errorf("fail to revert the change to `main.tf` file, err = %v", err)
		}
		return fmt.Errorf("fail to resize the instance, err = %v", err)
	}

	color.Green("Expanding file system ...")
	expand := "sudo growpart /dev/nvme0n1 1 && sudo resize2fs /dev/nvme0n1p1"
	return util.RemoteRun(name, expand, "ubuntu")
}
