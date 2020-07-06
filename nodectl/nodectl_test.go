package nodectl_test

import (
	"fmt"
	"os"
	"runtime"
	"testing/quick"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/darknode-cli/nodectl"
)

var _ = Describe("Nodectl", func() {

	Context("when deploying node on AWS", func() {
		node := fmt.Sprintf("aws-testing-%v", runtime.GOOS)

		Context("when the given params are valid", func() {
			Context("when using default values for the params", func() {
				It("should work without any error", func() {
					args := append(
						os.Args[0:1],
						CommandUp,
						flag("aws", ""),
						flag("name", node),
						flag("aws-access-key", os.Getenv("aws_access_key")),
						flag("aws-secret-key", os.Getenv("aws_secret_key")),
						flag("tags", "mainnet,testing"),
					)
					Expect(App().Run(args)).Should(Succeed())
				})

				It("show the details of the node info with list command", func() {
					name := os.Args[0:1]
					args := append(os.Args[0:1], CommandList)
					Expect(App().Run(args)).Should(Succeed())

					args = append(name, CommandList, flag("tags", "mainnet"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(name, CommandList, flag("tags", "testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(name, CommandList, flag("tags", "mainnet,testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(name, CommandList, flag("tags", "testnet"))
					err := App().Run(args)
					Expect(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring("cannot find any darknode with given tags"))
				})

				It("should be able to destroy the darknode", func() {
					args := append(os.Args[0:1], "down", "-force", node)
					Expect(App().Run(args)).Should(Succeed())
				})
			})
		})

		Context("when giving invalid params", func() {
			Context("when provider flag is not provided", func() {
				It("should return an error", func() {
					args := append(
						os.Args[0:1],
						CommandUp,
						flag("name", node),
						flag("aws-access-key", os.Getenv("aws_access_key")),
						flag("aws-secret-key", os.Getenv("aws_secret_key")),
						flag("tags", "mainnet,testing"),
					)
					err := App().Run(args)
					Expect(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring("unknown cloud provider"))
				})
			})

			Context("when name flag is not provided", func() {
				It("should return an error", func() {
					args := append(
						os.Args[0:1],
						CommandUp,
						flag("aws", ""),
						flag("aws-access-key", os.Getenv("aws_access_key")),
						flag("aws-secret-key", os.Getenv("aws_secret_key")),
						flag("tags", "mainnet,testing"),
					)
					err := App().Run(args)
					Expect(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring("node name cannot be empty"))
				})
			})

			Context("when giving an invalid network", func() {
				It("should return an error", func() {
					test := func(network string) bool {
						switch network {
						case "mainnet", "testnet", "devnet":
							return true
						default:
							args := append(
								os.Args[0:1],
								CommandUp,
								flag("aws", ""),
								flag("name", node),
								flag("network", network),
								flag("aws-access-key", os.Getenv("aws_access_key")),
								flag("aws-secret-key", os.Getenv("aws_secret_key")),
								flag("tags", "mainnet,testing"),
							)
							err := App().Run(args)
							Expect(err).ShouldNot(BeNil())
							Expect(err.Error()).Should(ContainSubstring("unknown network"))
							return true
						}
					}

					Expect(quick.Check(test, nil)).NotTo(HaveOccurred())
				})
			})

			Context("when not providing credentials", func() {
				It("should return an error", func() {
					test := func(accessKey, secretKey string) bool {
						args := append(
							os.Args[0:1],
							CommandUp,
							flag("aws", ""),
							flag("name", node),
							flag("tags", "mainnet,testing"),
						)
						err := App().Run(args)
						Expect(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring("invalid credentials"))
						return true
					}

					Expect(quick.Check(test, nil)).NotTo(HaveOccurred())
				})
			})

			Context("when providing invalid credentials", func() {
				It("should return an error", func() {
					test := func(accessKey, secretKey string) bool {
						args := append(
							os.Args[0:1],
							CommandUp,
							flag("aws", ""),
							flag("name", node),
							flag("aws-access-key", accessKey),
							flag("aws-secret-key", secretKey),
							flag("tags", "mainnet,testing"),
						)
						err := App().Run(args)
						Expect(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring("AuthFailure"))
						return true
					}

					Expect(quick.Check(test, &quick.Config{
						MaxCount: 10,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing invalid region", func() {
				It("should return an error", func() {
					test := func(region string) bool {
						if region == "" {
							return true
						}
						args := append(
							os.Args[0:1],
							CommandUp,
							flag("aws", ""),
							flag("name", node),
							flag("aws-access-key", os.Getenv("aws_access_key")),
							flag("aws-secret-key", os.Getenv("aws_secret_key")),
							flag("aws-region", region),
							flag("tags", "mainnet,testing"),
						)
						err := App().Run(args)
						Expect(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(" not in your available regions"))
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 10,
					})).NotTo(HaveOccurred())
				})
			})
		})
	})
})
