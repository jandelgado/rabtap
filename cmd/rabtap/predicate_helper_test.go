// Copyright (C) 2024Jan Delgado

package main

// a Predicate returning a constant value
type constantPred struct{ val bool }

func (s constantPred) Eval(_ map[string]interface{}) (bool, error) {
	return s.val, nil
}

// a predicate that delegates to a func
type funcPred struct {
	f func(map[string]interface{}) (bool, error)
}

func (s funcPred) Eval(env map[string]interface{}) (bool, error) {
	return s.f(env)
}
