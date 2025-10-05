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

// NewTreeNode returns a single node tree
func NewTreeNode(text string) *TreeNode {
	return &TreeNode{Text: text}
}

// HasChildren returns true if the node has children, otherwise false
func (s *TreeNode) HasChildren() bool {
	return len(s.Children) > 0
}

// Add adds a child node to the given node
func (s *TreeNode) Add(node *TreeNode) *TreeNode {
	node.Parent = s
	s.Children = append(s.Children, node)
	return node
}

// AddList adds a list of child node to the given node
func (s *TreeNode) AddList(nodes []*TreeNode) {
	for _, node := range nodes {
		s.Add(node)
	}
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
		switch {
		case parent.Parent == nil:
			// nop
		case parent.IsLastChild():
			treeLines = "   " + treeLines
		default:
			treeLines = "│  " + treeLines
		}
	}

	switch {
	case node.Parent == nil:
		// no treeLine for root element
	case node.IsLastChild():
		treeLines += "└─ "
	default:
		treeLines += "├─ "
	}
	_, _ = fmt.Fprintf(buffer, "%s%s\n", treeLines, node.Text)

	for _, p := range node.Children {
		PrintTree(p, buffer)
	}
}
