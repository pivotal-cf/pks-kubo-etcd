package deploy_test

import (
	"acceptance-tests/helpers"
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Multiple Instances", func() {
	var (
		etcdClientURLs []string
	)

	BeforeEach(func() {
		etcdClientURLs = bosh.GenerateAndSetDeploymentManifest(
			directorUUIDStub.Name(),
			helpers.InstanceCount1NodeStubPath,
			helpers.PersistentDiskStubPath,
			config.IAASSettingsStubPath,
			nameOverridesStub.Name(),
		)

		By("deploying")
		Expect(bosh.Command("-n", "deploy").Wait(helpers.DEFAULT_TIMEOUT)).To(Exit(0))

		Expect(len(etcdClientURLs)).To(Equal(1))
	})

	AfterEach(func() {
		By("delete deployment")
		Expect(bosh.Command("-n", "delete", "deployment", etcdName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
	})

	Describe("scaling from 1 node to 3", func() {
		It("succesfully scales to multiple etcd nodes", func() {
			for index, value := range etcdClientURLs {
				etcdClient := etcd.NewClient([]string{value})
				eatsKey := "eats-key" + string(index)
				eatsValue := "eats-value" + string(index)

				response, err := etcdClient.Create(eatsKey, eatsValue, 60)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())

				response, err = etcdClient.Get(eatsKey, false, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Node.Value).To(Equal(eatsValue))
			}

			etcdClientURLs = bosh.GenerateAndSetDeploymentManifest(
				directorUUIDStub.Name(),
				helpers.InstanceCount3NodesStubPath,
				helpers.PersistentDiskStubPath,
				config.IAASSettingsStubPath,
				nameOverridesStub.Name(),
			)

			By("deploying")
			Expect(bosh.Command("-n", "deploy").Wait(helpers.DEFAULT_TIMEOUT)).To(Exit(0))

			Expect(len(etcdClientURLs)).To(Equal(3))
			for index, value := range etcdClientURLs {
				etcdClient := etcd.NewClient([]string{value})

				eatsKey := fmt.Sprintf("eats-key%d", index)
				eatsValue := fmt.Sprintf("eats-value%d", index)

				response, err := etcdClient.Create(eatsKey, eatsValue, 60)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).ToNot(BeNil())
			}

			for _, value := range etcdClientURLs {
				etcdClient := etcd.NewClient([]string{value})

				for index, _ := range etcdClientURLs {
					eatsKey := fmt.Sprintf("eats-key%d", index)
					eatsValue := fmt.Sprintf("eats-value%d", index)

					response, err := etcdClient.Get(eatsKey, false, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Node.Value).To(Equal(eatsValue))
				}
			}
		})
	})
})
