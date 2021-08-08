// Copyright (C) 2017 Jan Delgado

package rabtap

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnescapeString(t *testing.T) {
	assert.Equal(t, "", unescapeStr(""))
	assert.Equal(t, "a", unescapeStr("a"))
	assert.Equal(t, "", unescapeStr("\\"))
	assert.Equal(t, "\\", unescapeStr("\\\\"))
	assert.Equal(t, "a:", unescapeStr("a\\:"))
}

func TestSplitExchangeAndBindingSimpleCaseProceeds(t *testing.T) {
	e, b, err := splitExchangeAndBinding("abc:def")
	assert.Nil(t, err)
	assert.Equal(t, "abc", e)
	assert.Equal(t, "def", b)
}

func TestSplitExchangeAndBindingHonorsEscapedColons(t *testing.T) {
	e, b, err := splitExchangeAndBinding("abc\\:xyz\\:123:def\\:jkl:")
	assert.Nil(t, err)
	assert.Equal(t, "abc:xyz:123", e)
	assert.Equal(t, "def:jkl:", b)
}

func TestSplitExchangeAndBindingRaisesErrorMissingKey(t *testing.T) {
	_, _, err := splitExchangeAndBinding("abcdef")
	assert.NotNil(t, err)
}

func TestNewTapConfigurationIsConstructedCorrecly(t *testing.T) {

	url, _ := url.Parse("uri")
	tc, err := NewTapConfiguration(url, "e1:b1,e2:b2")

	assert.Nil(t, err)
	assert.Equal(t, url, tc.AmqpURI)
	assert.Equal(t, 2, len(tc.Exchanges))
	assert.Equal(t, "e1", tc.Exchanges[0].Exchange)
	assert.Equal(t, "b1", tc.Exchanges[0].BindingKey)
	assert.Equal(t, "e2", tc.Exchanges[1].Exchange)
	assert.Equal(t, "b2", tc.Exchanges[1].BindingKey)
}

func TestFaultyTapConfiguration(t *testing.T) {

	url, _ := url.Parse("uri")
	_, err := NewTapConfiguration(url, "exchange")

	assert.NotNil(t, err)
}
