// Copyright (C) 2017 Jan Delgado

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasChildrenOfEmptyTreeReturnsFalse(t *testing.T) {
	tree := NewTreeNode("root")
	assert.False(t, tree.HasChildren())
}

func TestHasChildrenOfNonEmptyTreeReturnsTrue(t *testing.T) {
	tree := NewTreeNode("root")
	tree.Add(NewTreeNode("child"))
	assert.True(t, tree.HasChildren())
}

func TestAddMultipleNodes(t *testing.T) {
	tree := NewTreeNode("root")
	nodes := []*TreeNode{NewTreeNode("child1"), NewTreeNode("child2")}
	tree.AddList(nodes)
	assert.Equal(t, 2, len(tree.Children))
	assert.Equal(t, "child1", tree.Children[0].Text)
	assert.Equal(t, "child2", tree.Children[1].Text)
}

// ExampleNewInfoTree shows how to create a trees root-node
func ExampleNewTreeNode() {
	tree := NewTreeNode("root")

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	PrintTree(tree, writer)
	writer.Flush()
	fmt.Print(buffer.String())

	// Output:
	// root
}

// Example shows how to construct and print an tree with several nodes
func ExamplePrintTree() {

	tree := NewTreeNode("root")
	tree.Add(NewTreeNode("child1")).Add(NewTreeNode("child1.1"))
	tree.Add(NewTreeNode("child2"))
	tree.Add(NewTreeNode("child3")).Add(NewTreeNode("child4"))

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	PrintTree(tree, writer)
	writer.Flush()
	fmt.Print(buffer.String())

	// Output:
	// root
	// ├─ child1
	// │  └─ child1.1
	// ├─ child2
	// └─ child3
	//    └─ child4
}
