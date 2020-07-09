package provider

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type providerGCP struct {
	credFile string
}

type terraformGCP struct {
	Name           string
	CredentialFile string
	Project        string
	Zone           string
	MachineType    string
	ConfigPath     string
	PubKeyPath     string
	PriKeyPath     string
	ServiceFile    string
	LatestVersion  string
}

// NewGCP creates a GCP provider.
func NewGCP(ctx *cli.Context) (Provider, error) {
	credFile, err := filepath.Abs(ctx.String("gcp-credentials"))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(credFile); err != nil {
		return nil, err
	}
	// Verify the user has required permission for deploying a darknode
	if err := validatePermission(credFile); err != nil {
		return nil, err
	}
	return providerGCP{
		credFile: credFile,
	}, nil
}

// Name implements the `Provider` interface
func (p providerGCP) Name() string {
	return NameGcp
}

// Deploy implements the `Provider` interface
func (p providerGCP) Deploy(ctx *cli.Context) error {
	// validate params and generate the terraform config template from the params
	terraformConfig, err := p.validateParams(ctx)
	if err != nil {
		return err
	}
	// initialise all files for the Ren node
	t, err := template.New("gcp").Parse(templateGCP)
	if err != nil {
		return err
	}
	if err := initNode(ctx, terraformConfig, t); err != nil {
		return err
	}

	// deploy the node using terraform
	return terraformApply(ctx)
}

func validatePermission(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	creds, error := google.CredentialsFromJSON(ctx, data, compute.CloudPlatformScope)
	if error != nil {
		return err
	}
	service, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return err
	}
	rb := &cloudresourcemanager.TestIamPermissionsRequest{
		Permissions: []string{"compute.instances.create", "compute.networks.create", "compute.firewalls.create"}, ForceSendFields: nil, NullFields: nil,
	}
	resp, err := service.Projects.TestIamPermissions(creds.ProjectID, rb).Context(ctx).Do()
	if err != nil {
		return err
	}
	if resp.HTTPStatusCode != http.StatusOK {
		return errors.New("unable to check permission for the credential file")
	}
	if len(resp.Permissions) < 3 {
		return ErrInsufficientPermission
	}
	return nil
}

func (p providerGCP) validateParams(ctx *cli.Context) (*terraformGCP, error) {
	if err := validateCommonParams(ctx); err != nil {
		return nil, err
	}

	// Validate the name as GCP has special requirements for the name
	name := strings.TrimSpace(ctx.String("name"))
	reg := "^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"
	match, err := regexp.MatchString(reg, name)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidNodeNameForGCP
	}

	// Validate the zone and machine params
	zone, machine, project, err := p.validateRegionAndMachine(ctx)
	if err != nil {
		return nil, err
	}

	// Generate the terraform config file for the node.
	version, err := util.LatestStableRelease()
	if err != nil {
		return nil, err
	}
	return &terraformGCP{
		Name:           name,
		CredentialFile: p.credFile,
		Project:        project,
		Zone:           zone,
		MachineType:    machine,
		ConfigPath:     fmt.Sprintf("~/.darknode/darknodes/%v/config.json", name),
		PubKeyPath:     fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair.pub", name),
		PriKeyPath:     fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair", name),
		ServiceFile:    darknodeService,
		LatestVersion:  version,
	}, nil
}

func (p providerGCP) validateRegionAndMachine(ctx *cli.Context) (string, string, string, error) {
	region := strings.ToLower(strings.TrimSpace(ctx.String("gcp-region")))
	machine := strings.ToLower(strings.TrimSpace(ctx.String("gcp-machine")))

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Retrieve project ID and initialize the computer service
	fileName, err := filepath.Abs(ctx.String("gcp-credentials"))
	if err != nil {
		return "", "", "", err
	}
	projectID, err := projectID(c, fileName)
	if err != nil {
		return "", "", "", err
	}
	service, err := compute.NewService(c, option.WithCredentialsFile(fileName))
	if err != nil {
		return "", "", "", err
	}

	// Select a random zone for user if they don't provide one.
	var zone string
	if region == "" {
		zone, err = randomZone(c, service, projectID)
		if err != nil {
			return "", "", "", err
		}
	} else {
		zone, err = validateZone(c, service, projectID, region)
		if err != nil {
			return "", "", "", err
		}
	}

	// Validate if the machine type is available in the zone.
	return zone, machine, projectID, validateMachineType(c, service, projectID, zone, machine)
}

func projectID(ctx context.Context, fileName string) (string, error) {
	// Parse the credential file
	credFile, err := filepath.Abs(fileName)
	if err != nil {
		return "", err
	}
	jsonFile, err := os.Open(credFile)
	if err != nil {
		return "", err
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}

	// Get the project ID
	credentials, err := google.CredentialsFromJSON(ctx, data, compute.ComputeScope)
	if err != nil {
		return "", err
	}
	return credentials.ProjectID, nil
}

// Randomly pick a available zone
func randomZone(ctx context.Context, service *compute.Service, projectID string) (string, error) {
	zones := make([]string, 0)
	req := service.Zones.List(projectID)
	if err := req.Pages(ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			zones = append(zones, zone.Name)
		}
		return nil
	}); err != nil {
		return "", err
	}
	if len(zones) == 0 {
		return "", ErrNoAvailableRegion
	}

	return zones[rand.Intn(len(zones))], nil
}

func validateZone(ctx context.Context, service *compute.Service, projectID, input string) (string, error) {
	reg := regexp.MustCompile("^(?P<region>[a-z]+-[a-z]+[1-9])(-(?P<zone>[a-g]))?$")
	if !reg.MatchString(input) {
		return "", errors.New("invalid region name")
	}

	values := util.CaptureGroups("^(?P<region>[a-z]+-[a-z]+[1-9])(-(?P<zone>[a-g]))?$", input)
	zone := values["zone"]

	// If user doesn't provide a zone in the region, we only need to validate the
	// region and randomly select a zone in the region.
	if zone == "" {
		availableRegion, err := service.Regions.Get(projectID, input).Context(ctx).Do()
		if err != nil {
			return "", err
		}
		if len(availableRegion.Zones) == 0 {
			return "", ErrNoAvailableRegion
		}

		zone = path.Base(availableRegion.Zones[rand.Intn(len(availableRegion.Zones))])
		return zone, nil
	}

	// If user gives both region and zone, we only need to validate the zone.
	_, err := service.Zones.Get(projectID, input).Context(ctx).Do()
	return input, err
}

func validateMachineType(ctx context.Context, service *compute.Service, projectID, zone, machine string) error {
	res, err := service.MachineTypes.Get(projectID, zone, machine).Context(ctx).Do()
	if err != nil {
		return err
	}
	if res.Name != machine {
		return fmt.Errorf("%v type is not available in %v ", machine, zone)
	}
	return nil
}
