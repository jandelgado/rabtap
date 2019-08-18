package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveTemplate(t *testing.T) {
	type Info struct {
		Name string
	}
	args := Info{"Jan"}

	const tpl = "hello {{ .Name }}"

	result := resolveTemplate("test", tpl, args, map[string]interface{}{})
	assert.Equal(t, "hello Jan", result)
}
