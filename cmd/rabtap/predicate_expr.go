// (c) copyright 2024 Jan Delgado

package main

import (
	"errors"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// ExprPredicate is a Predicate that evaluates using the expr package
type ExprPredicate struct {
	prog *vm.Program
}

// NewExprPredicate creates a new predicate expression
func NewExprPredicate(exprstr string) (*ExprPredicate, error) {
	prog, err := expr.Compile(exprstr)
	if err != nil {
		return nil, err
	}
	return &ExprPredicate{prog: prog}, nil
}

// Eval evaluates the expression with a given set of parameters
func (s ExprPredicate) Eval(env map[string]interface{}) (bool, error) {
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
