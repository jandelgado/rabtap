// (c) copyright 2018 Jan Delgado

package main

// Predicate evaluates an expression to a boolean value
type Predicate interface {
	Eval(map[string]interface{}) (bool, error)
}
