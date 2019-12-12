package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1alpha1"
)

func invalidConfig(bridgeName string) nmstatev1alpha1.State {
	return nmstatev1alpha1.NewState(fmt.Sprintf(`interfaces:
  - name: %s
    type: linux-bridge
    state: invalid_state
`, bridgeName))
}

var _ = Describe("EnactmentCondition", func() {
	Context("when applying valid config", func() {
		BeforeEach(func() {
			By("Add some sleep time to vlan-filtering")
			runAtPods("cp", "/usr/local/bin/vlan-filtering", "/usr/local/bin/vlan-filtering.bak")
			runAtPods("sed", "-i", "$ a\\sleep 5", "/usr/local/bin/vlan-filtering")
			updateDesiredState(linuxBrUp(bridge1))
		})
		AfterEach(func() {
			By("Restore original vlan-filtering")
			runAtPods("mv", "/usr/local/bin/vlan-filtering.bak", "/usr/local/bin/vlan-filtering")
			updateDesiredState(linuxBrAbsent(bridge1))
			for _, node := range nodes {
				interfacesNameForNodeEventually(node).ShouldNot(ContainElement(bridge1))
			}
			By("Reset desired state at all nodes")
			resetDesiredStateForNodes()
		})
		It("should go from Progressing to Available", func() {
			progressConditions := []nmstatev1alpha1.Condition{
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionProgressing,
					Status: corev1.ConditionTrue,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionAvailable,
					Status: corev1.ConditionUnknown,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionFailing,
					Status: corev1.ConditionUnknown,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionMatching,
					Status: corev1.ConditionTrue,
				},
			}
			availableConditions := []nmstatev1alpha1.Condition{
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionProgressing,
					Status: corev1.ConditionFalse,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionAvailable,
					Status: corev1.ConditionTrue,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionFailing,
					Status: corev1.ConditionFalse,
				},
				nmstatev1alpha1.Condition{
					Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionMatching,
					Status: corev1.ConditionTrue,
				},
			}
			for _, node := range nodes {
				By("Check progressing state is reached")
				enactmentConditionsStatusEventually(node).Should(ConsistOf(progressConditions))

				By("Check available is the next condition")
				enactmentConditionsStatusEventually(node).Should(ConsistOf(availableConditions))

				By("Check that we available is keep")
				enactmentConditionsStatusConsistently(node).Should(ConsistOf(availableConditions))
			}
		})
	})

	Context("when applying invalid configuration", func() {
		BeforeEach(func() {
			updateDesiredState(invalidConfig(bridge1))

		})

		AfterEach(func() {
			By("Reset desired state at all nodes")
			resetDesiredStateForNodes()
		})

		It("should have Failing ConditionType set to true", func() {
			for _, node := range nodes {
				enactmentConditionsStatusEventually(node).Should(ConsistOf(
					nmstatev1alpha1.Condition{
						Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionFailing,
						Status: corev1.ConditionTrue,
					},
					nmstatev1alpha1.Condition{
						Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionAvailable,
						Status: corev1.ConditionFalse,
					},
					nmstatev1alpha1.Condition{
						Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionProgressing,
						Status: corev1.ConditionFalse,
					},
					nmstatev1alpha1.Condition{
						Type:   nmstatev1alpha1.NodeNetworkConfigurationEnactmentConditionMatching,
						Status: corev1.ConditionTrue,
					},
				))
			}
		})
	})
})
