package main_test

import (
	"os"
	"testing/quick"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/darknode-cli/darknode"
)

var _ = Describe("darknode", func() {

	Context("up command", func() {
		Context("when not specifying a provider tag", func() {
			It("should return an error", func() {
				args := append(os.Args[0:1], CmdUp, arg("name", "random"))
				ExpectErr(App(), args, "unknown cloud provider")
			})
		})
	})

	PContext("when deploying nodes on AWS", func() {
		Context("when the given params are valid", func() {
			Context("when using default values for most of the params", func() {
				It("should work without any error", func() {
					args := append(argsAWS, arg("name", nodeAWS))
					Expect(App().Run(args)).Should(Succeed())
				})

				It("show the details of the node info with list command", func() {
					bin := os.Args[0:1]
					args := append(bin, CmdList)
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet,testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testnet"))
					ExpectErr(App(), args, "cannot find any darknode with given tags")
				})

				It("should stop you from deploying another node with the same name", func() {
					args := append(argsAWS, arg("name", nodeAWS))
					ExpectErr(App(), args, "already exist")
				})

				It("should be able to destroy the darknode", func() {
					args := append(os.Args[0:1], CmdDown, "-force", nodeAWS)
					Expect(App().Run(args)).Should(Succeed())
				})
			})
		})

		Context("when giving invalid params", func() {
			Context("when name flag is not provided", func() {
				It("should return an error", func() {
					ExpectErr(App(), argsAWS, "node name cannot be empty")
				})
			})

			Context("when giving an invalid network", func() {
				It("should return an error", func() {
					test := func(network string) bool {
						switch network {
						case "mainnet", "testnet", "devnet", "":
						default:
							args := append(argsAWS, arg("name", nodeAWS), arg("network", network))
							ExpectErr(App(), args, "unknown network")
						}
						return true
					}

					Expect(quick.Check(test, nil)).NotTo(HaveOccurred())
				})
			})

			Context("when not providing credentials", func() {
				It("should return an error", func() {
					test := func(accessKey, secretKey string) bool {
						args := append(
							os.Args[0:1],
							CmdUp,
							arg("aws"),
							arg("name", nodeAWS),
							arg("tags", "mainnet,testing"),
						)
						ExpectErr(App(), args, "invalid credentials")
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
							CmdUp,
							arg("aws"),
							arg("name", nodeAWS),
							arg("aws-access-key", accessKey),
							arg("aws-secret-key", secretKey),
						)
						ExpectErr(App(), args, "AuthFailure")
						return true
					}

					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing invalid region", func() {
				It("should return an error", func() {
					test := func(region string) bool {
						if region == "" {
							return true
						}
						args := append(argsAWS, arg("name", nodeAWS), arg("aws-region", region))
						ExpectErr(App(), args, "not in your available regions")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing invalid instance type", func() {
				It("should return an error", func() {
					test := func(instance string) bool {
						if instance == "" {
							return true
						}
						args := append(argsAWS, arg("name", nodeAWS), arg("aws-instance", instance))
						ExpectErr(App(), args, "The following supplied instance types do not exist")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})
		})
	})

	PContext("when deploying nodes on Digital Ocean", func() {
		Context("when the given params are valid", func() {
			Context("when using default values for most of the params", func() {
				It("should work without any error", func() {
					args := append(argsDO, arg("name", nodeDO))
					Expect(App().Run(args)).Should(Succeed())
				})

				It("show the details of the node info with list command", func() {
					bin := os.Args[0:1]
					args := append(bin, CmdList)
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet,testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testnet"))
					ExpectErr(App(), args, "cannot find any darknode with given tags")
				})

				It("should stop you from deploying another node with the same name", func() {
					args := append(argsDO, arg("name", nodeDO))
					ExpectErr(App(), args, "already exist")
				})

				It("should be able to destroy the darknode", func() {
					args := append(os.Args[0:1], CmdDown, "-force", nodeDO)
					Expect(App().Run(args)).Should(Succeed())
				})
			})
		})

		Context("when giving invalid params", func() {
			Context("when name flag is not provided", func() {
				It("should return an error", func() {
					ExpectErr(App(), argsDO, "node name cannot be empty")
				})
			})

			Context("when giving an invalid network", func() {
				It("should return an error", func() {
					test := func(network string) bool {
						switch network {
						case "mainnet", "testnet", "devnet", "":
						default:
							args := append(argsDO, arg("name", nodeDO), arg("network", network))
							ExpectErr(App(), args, "unknown network")
						}
						return true
					}

					Expect(quick.Check(test, nil)).NotTo(HaveOccurred())
				})
			})

			Context("when providing invalid API token", func() {
				It("should return an error", func() {
					test := func(token string) bool {
						args := append(
							os.Args[0:1],
							CmdUp,
							arg("do"),
							arg("name", nodeAWS),
							arg("do-token", token),
						)
						ExpectErr(App(), args, "Unable to authenticate you")
						return true
					}

					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing an invalid region", func() {
				It("should return an error", func() {
					test := func(region string) bool {
						if region == "" {
							return true
						}
						args := append(argsDO, arg("name", nodeDO), arg("do-region", region))
						ExpectErr(App(), args, "is not in your available regions")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing an invalid droplet type", func() {
				It("should return an error", func() {
					test := func(droplet string) bool {
						if droplet == "" {
							return true
						}
						args := append(argsDO, arg("name", nodeDO), arg("do-droplet", droplet))
						ExpectErr(App(), args, "selected instance type is not available")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})
		})
	})

	Context("when deploying nodes on GCP", func() {
		Context("when the given params are valid", func() {
			Context("when using default values for most of the params", func() {
				It("should work without any error", func() {
					args := append(argsGCP, arg("name", nodeAWS))
					Expect(App().Run(args)).Should(Succeed())
				})

				It("show the details of the node info with list command", func() {
					bin := os.Args[0:1]
					args := append(bin, CmdList)
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "mainnet,testing"))
					Expect(App().Run(args)).Should(Succeed())

					args = append(bin, CmdList, arg("tags", "testnet"))
					ExpectErr(App(), args, "cannot find any darknode with given tags")
				})

				It("should stop you from deploying another node with the same name", func() {
					args := append(argsGCP, arg("name", nodeAWS))
					ExpectErr(App(), args, "already exist")
				})

				It("should be able to destroy the darknode", func() {
					args := append(os.Args[0:1], CmdDown, "-force", nodeAWS)
					Expect(App().Run(args)).Should(Succeed())
				})
			})
		})

		Context("when giving invalid params", func() {
			Context("when name flag is not provided", func() {
				It("should return an error", func() {
					ExpectErr(App(), argsGCP, "node name cannot be empty")
				})
			})

			Context("when giving an invalid network", func() {
				It("should return an error", func() {
					test := func(network string) bool {
						switch network {
						case "mainnet", "testnet", "devnet", "":
						default:
							args := append(argsGCP, arg("name", nodeGCP), arg("network", network))
							ExpectErr(App(), args, "unknown network")
						}
						return true
					}

					Expect(quick.Check(test, &quick.Config{MaxCount: 5})).NotTo(HaveOccurred())
				})
			})

			Context("when providing an invalid credential file", func() {
				It("should return an error", func() {
					test := func(cred string) bool {
						args := append(
							os.Args[0:1],
							CmdUp,
							arg("gcp"),
							arg("name", nodeGCP),
							arg("gcp-credentials", cred),
						)
						ExpectErr(App(), args, "no such file or directory")
						return true
					}

					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing an invalid region", func() {
				It("should return an error", func() {
					test := func(region string) bool {
						if region == "" {
							return true
						}
						args := append(argsGCP, arg("name", nodeGCP), arg("gcp-region", region))
						ExpectErr(App(), args, "invalid region name")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})

			Context("when providing an invalid machine type", func() {
				It("should return an error", func() {
					test := func(machine string) bool {
						if machine == "" {
							return true
						}
						args := append(argsGCP, arg("name", nodeGCP), arg("gcp-machine", machine))
						ExpectErr(App(), args, "Invalid value for field 'machineType'")
						return true
					}
					Expect(quick.Check(test, &quick.Config{
						MaxCount: 5,
					})).NotTo(HaveOccurred())
				})
			})
		})
	})
})
