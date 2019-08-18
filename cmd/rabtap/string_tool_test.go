package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterStringListOfEmptyLists(t *testing.T) {
	flags := []bool{}
	strs := []string{}
	assert.Equal(t, []string{}, filterStringList(flags, strs))
}

func TestFilterStringListOneElementKeptInList(t *testing.T) {
	flags := []bool{false, true, false}
	strs := []string{"A", "B", "C"}
	assert.Equal(t, []string{"B"}, filterStringList(flags, strs))
}
