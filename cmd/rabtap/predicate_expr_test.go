package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExprPredicateTrue(t *testing.T) {
	f, err := NewExprPredicate("1 == 1")
	require.NoError(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestExprPredicateFalse(t *testing.T) {
	f, err := NewExprPredicate("1 == 0")
	require.NoError(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.False(t, res)
}

func TestExprPredicateWithEnv(t *testing.T) {
	f, err := NewExprPredicate(`a == 1337 && b.X == 42 && c == "JD"`)
	require.NoError(t, err)
	params := make(map[string]interface{}, 1)
	params["a"] = 1337
	params["b"] = struct{ X int }{X: 42}
	params["c"] = "JD"
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestExprPredicateReturnsErrorOnInvalidSyntax(t *testing.T) {
	_, err := NewExprPredicate(")invalid syntax(")
	assert.ErrorContains(t, err, "unexpected token")
}

func TestExprPredicateReturnsErrorOnEvalError(t *testing.T) {
	f, err := NewExprPredicate("(1/a) == 1")
	require.NoError(t, err)
	_, err = f.Eval(nil)
	assert.ErrorContains(t, err, "invalid operation")
}
func TestExprPredicateReturnsErrorOnNonBoolReturnValue(t *testing.T) {
	f, err := NewExprPredicate("1+1")
	require.NoError(t, err)
	params := map[string]interface{}{}
	_, err = f.Eval(params)
	assert.ErrorContains(t, err, "expression does not evaluate to bool")
}
