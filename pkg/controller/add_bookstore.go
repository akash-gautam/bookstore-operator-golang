package controller

import (
	"bookstore-operator/pkg/controller/bookstore"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, bookstore.Add)
}
