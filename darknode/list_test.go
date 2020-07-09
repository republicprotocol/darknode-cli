package darknode_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/darknode-cli/darknode"
)

var _ = Describe("list command", func() {
	list := "list"

	Context("when there's no darknodes", func() {
		It("should return an error", func() {
			args := append(os.Args[0:1], list)
			err := App().Run(args)
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(ContainSubstring("cannot find any darknode with given tags"))
		})
	})
})
