package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruePredicate(t *testing.T) {
	res, err := TruePredicate.Eval(nil)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestPredicateTrue(t *testing.T) {
	f, err := NewPredicateExpression("1 == 1")
	assert.Nil(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestPredicateFalse(t *testing.T) {
	f, err := NewPredicateExpression("1 == 0")
	assert.Nil(t, err)
	params := map[string]interface{}{}
	res, err := f.Eval(params)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestPredicateWithParams(t *testing.T) {
	f, err := NewPredicateExpression("a == 1337 && b.X == 42")
	assert.Nil(t, err)
	params := make(map[string]interface{}, 1)
	params["a"] = 1337
	params["b"] = struct{ X int }{X: 42}
	res, err := f.Eval(params)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestPredicateReturnsErrorOnInvalidSyntax(t *testing.T) {
	_, err := NewPredicateExpression("invalid syntax")
	assert.NotNil(t, err)
}

func TestPredicateReturnsErrorOnEvalError(t *testing.T) {
	f, err := NewPredicateExpression("(1/a) == 1")
	assert.Nil(t, err)
	_, err = f.Eval(nil)
	assert.NotNil(t, err)
}
func TestPredicateReturnsErrorOnNonBoolReturnValue(t *testing.T) {
	f, err := NewPredicateExpression("1+1")
	assert.Nil(t, err)
	params := map[string]interface{}{}
	_, err = f.Eval(params)
	assert.NotNil(t, err)
}
