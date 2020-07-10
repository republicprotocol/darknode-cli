package main_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/darknode-cli/util"
	"github.com/urfave/cli/v2"
)

// Commands to be tested
var (
	CmdList = "list"

	CmdUp = "up"

	CmdDown = "down"
)

// Name of the nodes for different cloud providers to avoid conflicts.
var (
	nodeAWS = fmt.Sprintf("testing-aws-%v", runtime.GOOS)

	nodeDO = fmt.Sprintf("testing-do-%v", runtime.GOOS)

	nodeGCP = fmt.Sprintf("testing-gcp-%v", runtime.GOOS)
)

// Arguments template of different providers for reuse purpose
var (
	argsAWS = append(
		[]string{"darknode"},
		CmdUp,
		arg("aws"),
		arg("aws-access-key", os.Getenv("aws_access_key")),
		arg("aws-secret-key", os.Getenv("aws_secret_key")),
		arg("tags", "mainnet,testing"),
	)

	argsDO = append(
		[]string{"darknode"},
		CmdUp,
		arg("do"),
		arg("do-token", os.Getenv("do_token")),
		arg("tags", "mainnet,testing"),
	)

	argsGCP = append(
		[]string{"darknode"},
		CmdUp,
		arg("gcp"),
		arg("gcp-credentials", os.Getenv("gcp_credentials")),
		arg("tags", "mainnet,testing"),
	)
)

var _ = BeforeSuite(func() {

	// Verify the AWS credentials has been set up
	Expect(os.Getenv("aws_access_key")).ShouldNot(BeEmpty())
	Expect(os.Getenv("aws_secret_key")).ShouldNot(BeEmpty())

	// Verify the DO API token has been set up
	Expect(os.Getenv("do_token")).ShouldNot(BeEmpty())

	// Verify the path of GCP credential file has been set up
	Expect(os.Getenv("gcp_credentials")).ShouldNot(BeEmpty())

	// Verify the existence of the ~/.darknode/darknodes folder
	path := filepath.Join(os.Getenv("HOME"), ".darknode", "darknodes")
	_, err := os.Stat(path)
	Expect(err).NotTo(HaveOccurred())

	// Verify terraform has been installed
	Expect(util.Run("command", "-v", "terraform")).Should(Succeed())
})

func TestDarknodeCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Darknode-CLI Suite")
}

func arg(name string, values ...string) string {
	if len(values) == 0 {
		return fmt.Sprintf("--%v", name)
	} else {
		return fmt.Sprintf("--%v=%v", name, values[0])
	}
}

// Check if error returned by the app contains given text.
func ExpectErr(app *cli.App, args []string, text string) {
	err := app.Run(args)
	Expect(err).ShouldNot(BeNil())
	Expect(err.Error()).Should(ContainSubstring(text))
}
