package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatsAreConvertedIntoPercentageValues(t *testing.T) {
	f := rabtapTemplateFuncs{}

	assert.Equal(t, 0, f.toPercent(0.))
	assert.Equal(t, 99, f.toPercent(0.99))
	assert.Equal(t, 100, f.toPercent(0.999))
	assert.Equal(t, 100, f.toPercent(1.0))
}

func TestBoolIsConvertedToYesOrNo(t *testing.T) {
	f := rabtapTemplateFuncs{}

	assert.Equal(t, "no", f.asYesNo(false))
	assert.Equal(t, "yes", f.asYesNo(true))
}
