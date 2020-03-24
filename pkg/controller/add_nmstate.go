package controller

import (
	"github.com/nmstate/kubernetes-nmstate/pkg/controller/nmstate"
	"github.com/nmstate/kubernetes-nmstate/pkg/environment"
)

func init() {
	if !environment.IsOperator() {
		return
	}

	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nmstate.Add)
}
