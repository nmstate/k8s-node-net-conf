package helper

import (
	"fmt"

	"github.com/tidwall/gjson"

	yaml "sigs.k8s.io/yaml"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1alpha1"
)

func getBridgesUp(desiredState nmstatev1alpha1.State) (map[string][]string, error) {
	foundBridgesWithPorts := map[string][]string{}

	desiredStateYaml, err := yaml.YAMLToJSON([]byte(desiredState))
	if err != nil {
		return foundBridgesWithPorts, fmt.Errorf("error converting desiredState to JSON: %v", err)
	}

	bridgesUp := gjson.ParseBytes(desiredStateYaml).
		Get("interfaces.#(type==linux-bridge)#").
		Get("#(state==up)#").
		Array()

	for _, bridgeUp := range bridgesUp {
		portList := []string{}
		for _, port := range bridgeUp.Get("bridge.port.#.name").Array() {
			portList = append(portList, port.String())
		}

		foundBridgesWithPorts[bridgeUp.Get("name").String()] = portList
	}

	return foundBridgesWithPorts, nil
}
