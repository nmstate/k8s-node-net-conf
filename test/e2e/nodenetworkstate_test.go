package e2e

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	nmstatev1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1"
)

var _ = Describe("NodeNetworkState", func() {
	Context("when desiredState is configured", func() {
		Context("with a linux bridge", func() {
			var (
				br1Up = nmstatev1.State(`interfaces:
  - name: eth1
    type: ethernet
    state: up
  - name: br1
    type: linux-bridge
    state: up
    bridge:
      options:
        stp:
          enabled: false
      port:
        - name: eth1
          stp-hairpin-mode: false
          stp-path-cost: 100
          stp-priority: 32
`)
			)

			BeforeEach(func() {
				updateDesiredState(namespace, br1Up)
			})
			AfterEach(func() {

				// First we clean desired state if we
				// don't do that nmstate recreates the bridge
				resetDesiredStateForNodes(namespace)

				// TODO: Add status conditions to ensure that
				//       it has being really reset so we can
				//       remove this ugly sleep
				time.Sleep(1 * time.Second)

				// Let's clean the bridge directly in the node
				// bypassing nmstate
				deleteBridgeAtNodes("br1")
			})
			It("should have the linux bridge at currentState", func() {
				for _, node := range nodes {
					Eventually(func() []string {
						var currentStateYaml nmstatev1.State
						currentState(namespace, node, &currentStateYaml).ShouldNot(BeEmpty())

						interfaces := interfaces(currentStateYaml)
						Expect(interfaces).ToNot(BeEmpty(), "Node %s should have network interfaces", node)

						return interfacesName(interfaces)
					}, ReadTimeout, ReadInterval).Should(ContainElement("br1"))
				}
			})
		})
	})
})
