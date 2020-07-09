package darknode

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
	"github.com/renproject/darknode-cli/darknode/provider"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

// ErrInvalidInstanceSize is returned when the given instance size is invalid.
var (
	ErrInvalidInstanceSize = errors.New("invalid instance size")
)

// Regex for all the providers used for updating terraform config files.
var (
	InstanceAws = `instance_type\s+=\s*"(?P<instance>.+)"`

	InstanceDo = `size\s+=\s*"(?P<instance>.+)"`

	InstanceGcp = `machine_type\s+=\s+"(?P<instance>.+)"`
)

var (
	StorageAWS = `volume_size = `
)

type Change struct {
	Regex       string
	Replacement string
}

func resize(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	if err := util.ValidateNodeName(name); err != nil {
		return err
	}
	instance := ctx.String("instance")
	storage := ctx.Int("storage")

	// Fetch the cloud provider
	p, err := provider.GetProvider(name)
	if err != nil {
		return err
	}
	changes := []Change{}

	// Refresh terraform status
	color.Green("Refreshing darknode status")
	refresh := fmt.Sprintf("cd %v && terraform refresh", util.NodePath(name))
	if err := util.Run("bash", "-c", refresh); err != nil {
		return fmt.Errorf("cannot refresh terraform status, %v", err)
	}

	// Add storage change if user wants to expand the storage
	if storage != 0 {
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
		changes = append(changes, change)
	}

	// Add instance change if user wants to change the instance type
	if instance != "" {
		var change Change
		switch p {
		case provider.NameAws:
			change = Change{InstanceAws, fmt.Sprintf("volume_size = %v", storage)}
		case provider.NameDo:
			change = Change{InstanceDo, fmt.Sprintf(`size        = "%v"`, storage)}
		case provider.NameGcp:
			change = Change{InstanceDo, fmt.Sprintf(`machine_type = "%v"`, storage)}
		default:
			panic("unknown provider")
		}
		changes = append(changes, change)
	}

	// Apply the changes
	return applyChanges(name, changes)
}

func applyChanges(name string, changes []Change) error {
	// Read the config file
	path := filepath.Join(util.NodePath(name), "main.tf")
	tf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	original := make([]byte, len(tf))
	copy(original, tf)

	// Update the config file with the changes we want to make
	for _, change := range changes {
		reg, err := regexp.Compile(change.Regex)
		if err != nil {
			return err
		}

		tf = reg.ReplaceAll(tf, []byte(change.Replacement))
	}
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
	return nil
}
