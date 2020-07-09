package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	"github.com/digitalocean/godo"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

type providerDO struct {
	token string
}

type terraformDO struct {
	Name          string
	Token         string
	Region        string
	Size          string
	ConfigPath    string
	PubKeyPath    string
	PriKeyPath    string
	ServiceFile   string
	LatestVersion string
}

// NewDo creates a Digital Ocean provider.
func NewDo(ctx *cli.Context) (Provider, error) {
	token := ctx.String("do-token")

	return providerDO{
		token: token,
	}, nil
}

// Name implements the `Provider` interface
func (p providerDO) Name() string {
	return NameDo
}

// Deploy implements the `Provider` interface
func (p providerDO) Deploy(ctx *cli.Context) error {
	// validate params and generate the terraform config template from the params
	terraformConfig, err := p.validateParams(ctx)
	if err != nil {
		return err
	}

	// initialise all files for the Ren node
	t, err := template.New("do").Parse(templateDO)
	if err != nil {
		return err
	}
	if err := initNode(ctx, terraformConfig, t); err != nil {
		return err
	}

	// deploy the node using terraform
	return terraformApply(ctx)
}

func (p providerDO) validateParams(ctx *cli.Context) (*terraformDO, error) {
	name := ctx.String("name")
	if err := validateCommonParams(ctx); err != nil {
		return nil, err
	}
	region, err := p.validateRegion(ctx)
	if err != nil {
		return nil, err
	}
	droplet, err := p.validateDroplet(ctx, region)
	if err != nil {
		return nil, err
	}
	version, err := util.LatestStableRelease()
	if err != nil {
		return nil, err
	}
	return &terraformDO{
		Name:          name,
		Token:         p.token,
		Region:        region.Slug,
		Size:          droplet,
		ConfigPath:    fmt.Sprintf("~/.darknode/darknodes/%v/config.json", name),
		PubKeyPath:    fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair.pub", name),
		PriKeyPath:    fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair", name),
		ServiceFile:   darknodeService,
		LatestVersion: version,
	}, nil
}

func (p providerDO) validateRegion(ctx *cli.Context) (godo.Region, error) {
	region := ctx.String("do-region")
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch all available regions
	client := godo.NewFromToken(p.token)
	regions, response, err := client.Regions.List(c, nil)
	if err != nil {
		return godo.Region{}, err
	}
	if err := util.VerifyStatusCode(response.Response, http.StatusOK); err != nil {
		return godo.Region{}, err
	}
	if len(regions) == 0 {
		return godo.Region{}, ErrNoAvailableRegion
	}

	// Validate the given region or randomly pick one for the user
	if region == "" {
		return regions[rand.Intn(len(regions))], nil
	} else {
		all := ""
		for _, reg := range regions {
			all += reg.Slug + ","
			if reg.Slug == region {
				return reg, nil
			}
		}
		all = all[:len(all)-1]
		return godo.Region{}, fmt.Errorf("%v is not in your available regions [ %v ]", region, all)
	}
}

func (p providerDO) validateDroplet(ctx *cli.Context, region godo.Region) (string, error) {
	droplet := ctx.String("do-droplet")
	if !util.StringInSlice(droplet, region.Sizes) {
		return "", ErrInstanceTypeNotAvailable
	}
	return droplet, nil
}
