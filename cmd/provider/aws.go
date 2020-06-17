package provider

import (
	"errors"
	"log"
	"math/rand"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/renproject/darknode-cli/darknode"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli"
)

type providerAws struct {
	accessKey string
	secretKey string
}

func NewAws(ctx *cli.Context) (Provider, error) {
	accessKey := ctx.String("aws-access-key")
	secretKey := ctx.String("aws-secret-key")

	// Try reading the credential files if user does not provide credentials directly
	if accessKey == "" || secretKey == "" {
		cred := credentials.NewSharedCredentials("", ctx.String("aws-profile"))
		credValue, err := cred.Get()
		if err != nil {
			return nil, err
		}
		accessKey, secretKey = credValue.AccessKeyID, credValue.SecretAccessKey
		if accessKey == "" || secretKey == "" {
			return nil, err
		}
	}

	return providerAws{
		accessKey: accessKey,
		secretKey: secretKey,
	}, nil
}

func (p providerAws) Name() string {
	return NameAws
}

func (p providerAws) Deploy(ctx *cli.Context) error {
	name := ctx.String("name")
	tags := ctx.String("tags")

	latestVersion, err := util.LatestStableRelease()
	if err != nil {
		return err
	}
	region, instance, err := p.validateRegionAndInstance(ctx)
	if err != nil {
		return err
	}

	// Initialization
	network, err := darknode.NewNetwork(ctx.String("network"))
	if err != nil {
		return err
	}
	if err := initNode(name, tags, network); err != nil {
		return err
	}

	// Generate terraform config and start deploying
	if err := p.tfConfig(name, region, instance, latestVersion); err != nil {
		return err
	}
	if err := runTerraform(name); err != nil {
		return err
	}

	return outputURL(name)
}

func (p providerAws) validateRegionAndInstance(ctx *cli.Context) (string, string, error) {
	region := strings.ToLower(strings.TrimSpace(ctx.String("aws-region")))
	instance := strings.ToLower(strings.TrimSpace(ctx.String("aws-instance")))

	// Fetch valid regions for the user.
	cred := credentials.NewStaticCredentials(p.accessKey,p.secretKey,"")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: cred,
	})
	if err != nil {
		return "", "", err
	}
	service := ec2.New(sess)
	input := &ec2.DescribeRegionsInput{}
	result, err := service.DescribeRegions(input)
	if err != nil {
		return "", "", err
	}

	// Validate the given region or randomly pick one for the user
	if region == ""{
		randReg := result.Regions[rand.Intn(len(result.Regions))]
		region = *randReg.RegionName
	} else {
		valid := false
		for _, reg := range result.Regions {
			if *reg.RegionName == region {
				valid = true
				break
			}
		}
		if !valid {
			return "","", errors.New("invalid region")
		}
	}

	return region, instance, validateInstanceType(cred, region, instance)
}

func validateInstanceType(cred *credentials.Credentials, region, instance string) error {
	insSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: cred,
	})
	if err != nil {
		return err
	}
	service := ec2.New(insSession)
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []*string{aws.String(instance)},
	}
	result, err := service.DescribeInstanceTypes(input)
	if err != nil {
		return err
	}
	for _,res := range result.InstanceTypes {
		log.Print(*res.InstanceType)
	}
	return nil
}