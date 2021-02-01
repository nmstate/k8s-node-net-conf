package state

import (
	"os"

	"github.com/gobwas/glob"
	"github.com/nmstate/kubernetes-nmstate/api/shared"
	"github.com/nmstate/kubernetes-nmstate/pkg/environment"

	yaml "sigs.k8s.io/yaml"
)

var (
	interfacesFilterGlobFromEnv glob.Glob
)

func init() {
	if !environment.IsHandler() {
		return
	}
	interfacesFilter, isSet := os.LookupEnv("INTERFACES_FILTER")
	if !isSet {
		panic("INTERFACES_FILTER is mandatory")
	}
	interfacesFilterGlobFromEnv = glob.MustCompile(interfacesFilter)
}

func FilterOut(currentState shared.State) (shared.State, error) {
	return filterOut(currentState, interfacesFilterGlobFromEnv)
}

func filterOutRoutes(kind string, state map[string]interface{}, interfacesFilterGlob glob.Glob) {
	routesRaw, hasRoutes := state["routes"]
	if !hasRoutes {
		return
	}

	routes, ok := routesRaw.(map[string]interface{})
	if !ok {
		return
	}

	routesByKind := routes[kind].([]interface{})

	if routesByKind == nil {
		return
	}

	filteredRoutes := []interface{}{}
	for _, route := range routesByKind {
		name := route.(map[string]interface{})["next-hop-interface"]
		if !interfacesFilterGlob.Match(name.(string)) {
			filteredRoutes = append(filteredRoutes, route)
		}
	}

	state["routes"].(map[string]interface{})[kind] = filteredRoutes
}

func filterOutDynamicAttributes(iface map[string]interface{}) {
	// The gc-timer and hello-time are deep into linux-bridge like this
	//    - bridge:
	//        options:
	//          gc-timer: 13715
	//          hello-timer: 0
	if iface["type"] != "linux-bridge" {
		return
	}

	bridgeRaw, hasBridge := iface["bridge"]
	if !hasBridge {
		return
	}
	bridge, ok := bridgeRaw.(map[string]interface{})
	if !ok {
		return
	}

	optionsRaw, hasOptions := bridge["options"]
	if !hasOptions {
		return
	}
	options, ok := optionsRaw.(map[string]interface{})
	if !ok {
		return
	}

	delete(options, "gc-timer")
	delete(options, "hello-timer")
}

func filterOutInterfaces(state map[string]interface{}, interfacesFilterGlob glob.Glob) {
	interfaces := state["interfaces"]
	filteredInterfaces := []interface{}{}

	for _, iface := range interfaces.([]interface{}) {
		name := iface.(map[string]interface{})["name"]
		if !interfacesFilterGlob.Match(name.(string)) {
			filterOutDynamicAttributes(iface.(map[string]interface{}))
			filteredInterfaces = append(filteredInterfaces, iface)
		}
	}
	state["interfaces"] = filteredInterfaces
}

func filterOut(currentState shared.State, interfacesFilterGlob glob.Glob) (shared.State, error) {
	var state map[string]interface{}
	err := yaml.Unmarshal(currentState.Raw, &state)
	if err != nil {
		return currentState, err
	}

	filterOutInterfaces(state, interfacesFilterGlob)
	filterOutRoutes("running", state, interfacesFilterGlob)
	filterOutRoutes("config", state, interfacesFilterGlob)

	filteredState, err := yaml.Marshal(state)
	if err != nil {
		return currentState, err
	}

	return shared.NewState(string(filteredState)), nil
}
