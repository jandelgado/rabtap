// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTapConfiguration(t *testing.T) {

	tc, err := NewTapConfiguration("uri", "e1:b1,e2:b2")
	assert.Nil(t, err)
	assert.Equal(t, "uri", tc.AmqpURI)
	assert.Equal(t, 2, len(tc.Exchanges))
	assert.Equal(t, "e1", tc.Exchanges[0].Exchange)
	assert.Equal(t, "b1", tc.Exchanges[0].BindingKey)
	assert.Equal(t, "e2", tc.Exchanges[1].Exchange)
	assert.Equal(t, "b2", tc.Exchanges[1].BindingKey)
}

func TestFaultyTapConfiguration(t *testing.T) {

	_, err := NewTapConfiguration("uri", "exchange")

	assert.NotNil(t, err)
}
