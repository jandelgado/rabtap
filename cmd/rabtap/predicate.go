// (c) copyright 2018 Jan Delgado

package main

import (
	"errors"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Predicate evaluates an expression to a boolean value
type Predicate interface {
	Eval(map[string]interface{}) (bool, error)
}

// TruePredicate always evaluates to true
var TruePredicate = truePredicate{}

type truePredicate struct{}

func (s truePredicate) Eval(params map[string]interface{}) (bool, error) {
	return true, nil
}

// PredicateExpression implements an predicate expression evaluator using
// the govaluate package
type PredicateExpression struct {
	prog *vm.Program
}

// NewPredicateExpression creates a new predicate expression
func NewPredicateExpression(exprstr string) (Predicate, error) {
	prog, err := expr.Compile(exprstr)
	if err != nil {
		return nil, err
	}
	return &PredicateExpression{prog: prog}, nil
}

// Eval evaluates the expression with a given set of parameters
func (s PredicateExpression) Eval(env map[string]interface{}) (bool, error) {
	result, err := expr.Run(s.prog, env)
	if err != nil {
		return false, err
	}

	res, ok := result.(bool)
	if !ok {
		return false, errors.New("expression does not evaluate to bool")
	}
	return res, nil
}
