package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

type providerAWS struct {
	accessKey string
	secretKey string
}

type terraformAWS struct {
	Name          string
	Region        string
	InstanceType  string
	ConfigPath    string
	PubKeyPath    string
	PriKeyPath    string
	AccessKey     string
	SecretKey     string
	ServiceFile   string
	LatestVersion string
}

// NewAWS creates a AWS provider.
func NewAWS(ctx *cli.Context) (Provider, error) {
	accessKey := ctx.String("aws-access-key")
	secretKey := ctx.String("aws-secret-key")

	// Try reading the default credential file if user does not provide credentials directly
	if accessKey == "" || secretKey == "" {
		cred := credentials.NewSharedCredentials("", ctx.String("aws-profile"))
		credValue, err := cred.Get()
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
		accessKey, secretKey = credValue.AccessKeyID, credValue.SecretAccessKey
		if accessKey == "" || secretKey == "" {
			return nil, errors.New("invalid credentials")
		}
	}

	return providerAWS{
		accessKey: accessKey,
		secretKey: secretKey,
	}, nil
}

// Name implements the `Provider` interface
func (p providerAWS) Name() string {
	return NameAws
}

// Deploy implements the `Provider` interface
func (p providerAWS) Deploy(ctx *cli.Context) error {
	// validate params and generate the terraform config template from the params
	terraformConfig, err := p.validateParams(ctx)
	if err != nil {
		return err
	}

	// initialise all files for the Ren node
	t, err := template.New("aws").Parse(templateAWS)
	if err != nil {
		return err
	}
	if err := initNode(ctx, terraformConfig, t); err != nil {
		return err
	}

	// deploy the node using terraform
	return terraformApply(ctx)
}

func (p providerAWS) validateParams(ctx *cli.Context) (*terraformAWS, error) {
	name := ctx.String("name")
	if err := validateCommonParams(ctx); err != nil {
		return nil, err
	}
	region, err := p.validateRegion(ctx)
	if err != nil {
		return nil, err
	}
	instance, err := p.validateInstanceType(ctx, region)
	if err != nil {
		return nil, err
	}
	version, err := util.LatestStableRelease()
	if err != nil {
		return nil, err
	}

	return &terraformAWS{
		Name:          name,
		Region:        region,
		InstanceType:  instance,
		ConfigPath:    fmt.Sprintf("~/.darknode/darknodes/%v/config.json", name),
		PubKeyPath:    fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair.pub", name),
		PriKeyPath:    fmt.Sprintf("~/.darknode/darknodes/%v/ssh_keypair", name),
		AccessKey:     p.accessKey,
		SecretKey:     p.secretKey,
		ServiceFile:   darknodeService,
		LatestVersion: version,
	}, nil
}

func (p providerAWS) validateRegion(ctx *cli.Context) (string, error) {
	// Fetch all available regions for the user
	cred := credentials.NewStaticCredentials(p.accessKey, p.secretKey, "")
	region := strings.ToLower(strings.TrimSpace(ctx.String("aws-region")))
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: cred,
	})
	service := ec2.New(sess)
	input := &ec2.DescribeRegionsInput{}
	result, err := service.DescribeRegions(input)
	if err != nil {
		return "", err
	}
	if len(result.Regions) == 0 {
		return "", ErrNoAvailableRegion
	}
	// Validate the given region or randomly pick one for the user
	if region == "" {
		randReg := result.Regions[rand.Intn(len(result.Regions))]
		region = *randReg.RegionName
	} else {
		valid := false
		all := ""
		for _, reg := range result.Regions {
			all += *reg.RegionName + ","
			if *reg.RegionName == region {
				valid = true
				break
			}
		}
		all = all[:len(all)-1]
		if !valid {
			return "", fmt.Errorf("%v is not in your available regions [ %v ]", region, all)
		}
	}
	return region, nil
}

func (p providerAWS) validateInstanceType(ctx *cli.Context, region string) (string, error) {
	instance := strings.ToLower(strings.TrimSpace(ctx.String("aws-instance")))
	cred := credentials.NewStaticCredentials(p.accessKey, p.secretKey, "")
	insSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: cred,
	})
	if err != nil {
		return "", err
	}
	service := ec2.New(insSession)
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []*string{aws.String(instance)},
	}
	result, err := service.DescribeInstanceTypes(input)
	if err != nil {
		return "", err
	}
	valid := false
	for _, res := range result.InstanceTypes {
		if *res.InstanceType == instance {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("instance type %v is not available in region %v", instance, region)
	}
	return instance, nil
}
