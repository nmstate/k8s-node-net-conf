package controllers

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error messages formatting", func() {
	var (
		debugInvalidInput      = errors.New("      2021-04-15 08:17:35,057 root         DEBUG    Async action: Create checkpoint started\n      2021-04-15 08:17:35,060 root         DEBUG    Checkpoint None created for all devices\n")
		debugValidInput        = errors.New("A message containing debug that should not be removed.\n")
		tracebackInvalidInput  = errors.New("      Traceback (most recent call last):\n        File \"/usr/bin/nmstatectl\", line 11, in <module>\n          load_entry_point('nmstate==0.3.6', 'console_scripts', 'nmstatectl')()\n      File \"/usr/lib/python3.6/site-packages/nmstatectl/nmstatectl.py\", line 69, in main\n     f\"Interface {iface.name} has unknown slave: \"\n      libnmstate.error.NmstateValueError: Interface bond1 has unknown slave: eth10\n      ")
		tracebackInvalidOutput = "      libnmstate.error.NmstateValueError\n  Interface bond1 has unknown slave\n    eth10\n"
		tracebackValidInput    = errors.New("A message containing File \"/usr/bin/nmstatectl\" that should not be removed.\n")
		failedToExecuteInput   = errors.New(" failed to execute nmstatectl set --no-commit --timeout 480: 'exit status 1' '' '2021-02-22 11:10:08,962 root         WARNING  libnm version 1.26.7 mismatches NetworkManager version 1.29.9\n")
		failedToExecuteOutput  = "failed to execute nmstatectl set --no-commit --timeout 480: 'exit status 1'\n"
		pingInvalidInput       = errors.New("rolling back desired state configuration: failed runnig probes after network changes: failed runnig probe 'ping' with after network reconfiguration -> currentState: ---\n      dns-resolver:\n      config:\n          search: []\n          server: []\n        running: {}\n      route-rules:\n        config: []\n      : failed to retrieve default gw at runProbes: timed out waiting for the condition")
		pingInvalidOutput      = "      \n  failed to retrieve default gw at runProbes\n    timed out waiting for the condition\n"
		pingValidInput         = errors.New("rolling back desired state configuration: failed runnig probes after network changes: failed runnig probe 'ping' with after network reconfiguration.\nThe rest of the message should be kept.\n")
		pingValidOutput        = "rolling back desired state configuration\n  failed runnig probes after network changes\n    failed runnig probe 'ping' with after network reconfiguration.\nThe rest of the message should be kept.\n"
	)

	Context("With DEBUG text", func() {
		It("Should remove DEBUG message", func() {
			Expect(formatErrorString(debugInvalidInput)).To(Equal(""))
		})
		It("Should keep message with debug keyword", func() {
			Expect(formatErrorString(debugValidInput)).To(Equal(debugValidInput.Error()))
		})
	})

	Context("With Traceback text", func() {
		It("Should remove python traceback", func() {
			Expect(formatErrorString(tracebackInvalidInput)).To(Equal(tracebackInvalidOutput))
		})
		It("Should keep message with File keyword", func() {
			Expect(formatErrorString(tracebackValidInput)).To(Equal(tracebackValidInput.Error()))
		})
	})

	Context("With failed to execute text", func() {
		It("Should remove warning form the line", func() {
			Expect(formatErrorString(failedToExecuteInput)).To(Equal(failedToExecuteOutput))
		})
	})

	Context("With network reconfiguration text", func() {
		It("Should remove yaml", func() {
			Expect(formatErrorString(pingInvalidInput)).To(Equal(pingInvalidOutput))
		})
		It("Should keep message", func() {
			Expect(formatErrorString(pingValidInput)).To(Equal(pingValidOutput))
		})
	})
})
