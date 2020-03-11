package controller

import (
	"github.com/sunny0826/maoxian-operator/pkg/controller/maoxianbot"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, maoxianbot.Add)
}
