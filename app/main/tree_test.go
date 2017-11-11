// Copyright (C) 2017 Jan Delgado

package main

import (
	"bufio"
	"bytes"
	"fmt"
)

// ExampleNewInfoTree shows how to create a trees root-node
func ExampleNewInfoTree() {
	tree := NewInfoTree("root")

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

	tree := NewInfoTree("root")
	tree.AddChild("child1").AddChild("child1.1")
	tree.AddChild("child2")
	tree.AddChild("child3").AddChild("child4")

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	PrintTree(tree, writer)
	writer.Flush()
	fmt.Print(buffer.String())

	// Output:
	// root
	// ├── child1
	// │   └── child1.1
	// ├── child2
	// └── child3
	//     └── child4
}
