package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPredicateTrue(t *testing.T) {
	f, err := NewPredicateExpression("1 == 1")
	require.NoError(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestPredicateFalse(t *testing.T) {
	f, err := NewPredicateExpression("1 == 0")
	require.NoError(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.False(t, res)
}

func TestPredicateWithEnv(t *testing.T) {
	f, err := NewPredicateExpression(`a == 1337 && b.X == 42 && c == "JD"`)
	require.NoError(t, err)
	params := make(map[string]interface{}, 1)
	params["a"] = 1337
	params["b"] = struct{ X int }{X: 42}
	params["c"] = "JD"
	res, err := f.Eval(params)
	require.NoError(t, err)
	assert.True(t, res)
}

func TestPredicateReturnsErrorOnInvalidSyntax(t *testing.T) {
	_, err := NewPredicateExpression(")invalid syntax(")
	assert.ErrorContains(t, err, "unexpected token")
}

func TestPredicateReturnsErrorOnEvalError(t *testing.T) {
	f, err := NewPredicateExpression("(1/a) == 1")
	require.NoError(t, err)
	_, err = f.Eval(nil)
	assert.ErrorContains(t, err, "invalid operation")
}
func TestPredicateReturnsErrorOnNonBoolReturnValue(t *testing.T) {
	f, err := NewPredicateExpression("1+1")
	require.NoError(t, err)
	params := map[string]interface{}{}
	_, err = f.Eval(params)
	assert.ErrorContains(t, err, "expression does not evaluate to bool")
}
