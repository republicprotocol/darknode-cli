package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/renproject/darknode-cli/renvm"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

var (
	// ErrUnknownProvider is returned when user tries to deploy a darknode with an unknown cloud provider.
	ErrUnknownProvider = errors.New("unknown cloud provider")

	// ErrUnsupportedInstanceType is returned when the selected instance type cannot be created to user account.
	ErrInstanceTypeNotAvailable = errors.New("selected instance type is not available")

	// ErrRegionNotAvailable is returned when the selected region is not available to user account.
	ErrRegionNotAvailable = errors.New("selected region is not available")

	// ErrNoAvailableRegion is returned when the user account has no available regions.
	ErrNoAvailableRegion = errors.New("cannot find any available region with given account")

	// ErrInsufficientPermission is returned when user account to the required permission to create a VM instance.
	ErrInsufficientPermission = errors.New("insufficient permissions")

	// ErrInvalidNodeNameForGCP is returned the name doesn't meet the GCP requirements
	ErrInvalidNodeNameForGCP = errors.New("for google cloud, name must start with a lowercase letter followed by up to 62 lowercase letters, numbers, or hyphens, and cannot end with a hyphen")
)

var (
	NameAws = "aws"
	NameDo  = "do"
	NameGcp = "gcp"
)

var darknodeService = `[Unit]
Description=RenVM Darknode Daemon
AssertPathExists=$HOME/.darknode

[Service]
WorkingDirectory=$HOME/.darknode
ExecStart=$HOME/.darknode/bin/darknode --config $HOME/.darknode/config.json
LimitNOFILE=4096
Restart=on-failure
PrivateTmp=true
NoNewPrivileges=true

# Specifies which signal to use when killing a service. Defaults to SIGTERM.
# SIGHUP gives parity time to exit cleanly before SIGKILL (default 90s)
KillSignal=SIGHUP

[Install]
WantedBy=default.target`

type Provider interface {

	// Name of the Provider
	Name() string

	// Deploy the Ren node with from this provider
	Deploy(ctx *cli.Context) error
}

func ParseProvider(ctx *cli.Context) (Provider, error) {
	if ctx.Bool(NameAws) {
		return NewAWS(ctx)
	}

	if ctx.Bool(NameDo) {
		return NewDo(ctx)
	}

	if ctx.Bool(NameGcp) {
		return NewGCP(ctx)
	}

	return nil, ErrUnknownProvider
}

// Provider returns the provider of a darknode instance.
func GetProvider(name string) (string, error) {
	if name == "" {
		return "", util.ErrEmptyName
	}

	cmd := fmt.Sprintf("cd %v && terraform output provider", util.NodePath(name))
	provider, err := util.CommandOutput(cmd)
	return strings.TrimSpace(provider), err
}

// Validate the params which are available for all providers.
func validateCommonParams(ctx *cli.Context) error {
	// Validate common params
	name := ctx.String("name")
	if name == "" {
		return util.ErrEmptyName
	}
	if _, err := os.Stat(util.NodePath(name)); err == nil {
		return fmt.Errorf("node [%v] already exist", name)
	}
	_, err := renvm.NewNetwork(ctx.String("network"))
	if err != nil {
		return err
	}

	// Verify the config file if user wants to use their own config
	configFile := ctx.String("config")
	if configFile != "" {
		// verify the config exist and of the right format
		path, err := filepath.Abs(configFile)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err != nil {
			return errors.New("config file doesn't exist")
		}
		jsonFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer jsonFile.Close()
		var config renvm.Config
		if err := json.NewDecoder(jsonFile).Decode(&config); err != nil {
			return fmt.Errorf("incompatible config, err = %v", err)
		}
	}
	return nil
}

// Initialize files for deploying a darknode
func initNode(ctx *cli.Context, data interface{}, temp *template.Template) error {
	name := ctx.String("name")
	path := util.NodePath(name)

	// Create directory for the Ren node
	if err := os.Mkdir(path, 0700); err != nil {
		return err
	}

	// Create `tags.out` file
	tags := []byte(strings.TrimSpace(ctx.String("tags")))
	tagsPath := filepath.Join(path, "tags.out")
	if err := ioutil.WriteFile(tagsPath, tags, 0600); err != nil {
		return err
	}

	// Create `ssh_keypair` and `ssh_keypair.pub` files for the remote instance
	if err := util.GenerateSshKeyAndWriteToDir(name); err != nil {
		return err
	}

	// Create `config.json` for the node
	if err := initConfig(ctx); err != nil {
		return err
	}

	// Create `main.tf` file
	tfFile, err := os.Create(filepath.Join(path, "main.tf"))
	if err != nil {
		return err
	}
	return temp.Execute(tfFile, data)
}

func initConfig(ctx *cli.Context) error {
	name := ctx.String("name")
	configFile := ctx.String("config")
	destination := filepath.Join(util.NodePath(name), "config.json")

	if configFile != "" {
		path, err := filepath.Abs(configFile)
		if err != nil {
			return err
		}
		input, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(destination, input, 0600)
	} else {
		network, _ := renvm.NewNetwork(ctx.String("network"))
		config, err := renvm.NewConfig(network)
		if err != nil {
			return err
		}
		configData, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(destination, configData, 0600)
	}
}

func terraformApply(ctx *cli.Context) error {
	name := ctx.String("name")
	path := util.NodePath(name)

	init := fmt.Sprintf("cd %v && terraform init", path)
	if err := util.Run("bash", "-c", init); err != nil {
		return err
	}

	color.Green("Deploying darknode ... ")
	apply := fmt.Sprintf("cd %v && terraform apply -auto-approve -no-color", path)
	if err := util.Run("bash", "-c", apply); err != nil {
		return err
	}

	// Check if the deployment is successful as the exit code of terraform is not reliable
	// See https://github.com/hashicorp/terraform/issues/20671
	if _, err := util.IP(name); err != nil {
		return err
	}

	// Redirect to the registration link
	url, err := util.RegisterUrl(name)
	if err != nil {
		return err
	}
	color.Green("")
	color.Green("Congratulations! Your Darknode is deployed.")
	color.Green("Join the network by registering your Darknode at %s", url)
	return util.OpenInBrowser(url)
}
