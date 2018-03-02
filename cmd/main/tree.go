// Copyright (C) 2017 Jan Delgado

package main

// a simple tree strcture that can be printed on the console.

import (
	"fmt"
	"io"
)

// TreeNode is a node of the tree containing information to be displayed
type TreeNode struct {
	Text     string
	Parent   *TreeNode
	Children []*TreeNode
}

// NewInfoTree returns a new Info treee with just an root element
func NewInfoTree(text string) *TreeNode {
	return &TreeNode{Text: text}
}

// AddChild adds a child node to the given node and returns the newly
// created node.
func (s *TreeNode) AddChild(text string) *TreeNode {
	node := &TreeNode{Text: text}
	node.Parent = s
	s.Children = append(s.Children, node)
	return node
}

// IsLastChild returns true if the node is the last child node of it's parent
func (s *TreeNode) IsLastChild() bool {
	parent := s.Parent

	for i, p := range parent.Children {
		if p == s && i == len(parent.Children)-1 {
			return true
		}
	}
	return false
}

// PrintTree prints nicely formatted a tree structure
// TODO externalize treeLine symbols + as param
func PrintTree(node *TreeNode, buffer io.Writer) {
	treeLines := ""
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Parent == nil {

		} else if parent.IsLastChild() {
			treeLines = "    " + treeLines
		} else {
			treeLines = "│   " + treeLines
		}
	}

	if node.Parent == nil {
		// no treeLine for root element
	} else if node.IsLastChild() {
		treeLines = treeLines + "└── "
	} else {
		treeLines = treeLines + "├── "
	}
	fmt.Fprintf(buffer, "%s%s\n", treeLines, node.Text)

	for _, p := range node.Children {
		PrintTree(p, buffer)
	}
}
