package darknode_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	CommandList = "list"

	CommandUp = "up"
)

var _ = BeforeSuite(func() {
	Expect(os.Getenv("aws_access_key")).ShouldNot(BeEmpty())
	Expect(os.Getenv("aws_secret_key")).ShouldNot(BeEmpty())
})

func TestDarknode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Darknode Suite")
}

func flag(name, value string) string {
	if value == "" {
		return fmt.Sprintf("--%v", name)
	} else {
		return fmt.Sprintf("--%v=%v", name, value)
	}
}
