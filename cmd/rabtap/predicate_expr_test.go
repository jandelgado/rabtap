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

func TestExprPredicateWithInitalEnv(t *testing.T) {
	initEnv := map[string]interface{}{"a": 1337}
	// note that all variables are prefixed with "r" in the Eval() method
	f, err := NewExprPredicateWithEnv(`r.b < r.a`, initEnv)
	require.NoError(t, err)

	env := map[string]interface{}{"b": 100}
	res, err := f.Eval(env)

	require.NoError(t, err)
	assert.True(t, res)
}
func TestExprPredicateWithEvalEnv(t *testing.T) {
	f, err := NewExprPredicate(`r.a == 1337 && r.b.X == 42 && r.c == "JD"`)
	require.NoError(t, err)
	env := map[string]interface{}{
		"a": 1337,
		"b": struct{ X int }{X: 42},
		"c": "JD",
	}

	res, err := f.Eval(env)

	require.NoError(t, err)
	assert.True(t, res)
}

func TestExprPredicateReturnsErrorOnInvalidSyntax(t *testing.T) {
	_, err := NewExprPredicate(")invalid syntax(")
	assert.ErrorContains(t, err, "unexpected token")
}

func TestExprPredicateReturnsErrorOnEvalError(t *testing.T) {
	f, err := NewExprPredicate("(1/r.a) == 1")
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
