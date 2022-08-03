// Copyright (c) 2016 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// copied from https://github.com/dnaeon/go-dependency-graph-algorithm

package fixtures

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

// Node represents a single node in the graph with it's dependencies
type Node struct {
	// Name of the node
	name string

	typ interface{}

	// Dependencies of the node
	deps []*Node

	tag tag
}

// NewNode creates a new node
func NewNode(name string, typ interface{}, deps ...*Node) *Node {
	n := &Node{
		name: name,
		typ:  typ,
		deps: deps,
	}

	return n
}

type Graph []*Node

// Resolves the dependency graph
func resolveGraph(graph Graph) (Graph, error) {
	// A map containing the node names and the actual node object
	nodeNames := make(map[string]*Node)

	// A map containing the nodes and their dependencies
	nodeDependencies := make(map[string]mapset.Set)

	// Populate the maps
	for _, node := range graph {
		nodeNames[node.name] = node

		dependencySet := mapset.NewSet()
		for _, dep := range node.deps {
			dependencySet.Add(dep.name)
		}
		nodeDependencies[node.name] = dependencySet
	}

	// Iteratively find and remove nodes from the graph which have no dependencies.
	// If at some point there are still nodes in the graph and we cannot find
	// nodes without dependencies, that means we have a circular dependency
	var resolved Graph
	for len(nodeDependencies) != 0 {
		// Get all nodes from the graph which have no dependencies
		readySet := mapset.NewSet()
		for name, deps := range nodeDependencies {
			if deps.Cardinality() == 0 {
				readySet.Add(name)
			}
		}

		// If there aren't any ready nodes, then we have a cicular dependency
		if readySet.Cardinality() == 0 {
			var g Graph
			for name := range nodeDependencies {
				g = append(g, nodeNames[name])
			}

			return g, fmt.Errorf("circular dependency found")
		}

		// Remove the ready nodes and add them to the resolved graph
		for name := range readySet.Iter() {
			delete(nodeDependencies, name.(string))
			resolved = append(resolved, nodeNames[name.(string)])
		}

		// Also make sure to remove the ready nodes from the
		// remaining node dependencies as well
		for name, deps := range nodeDependencies {
			diff := deps.Difference(readySet)
			nodeDependencies[name] = diff
		}
	}

	return resolved, nil
}

// func (g Graph) index(name string) bool {

// }

func (g Graph) String() string {
	return strings.Join(g.toStrings(), ",")
}

func (g Graph) toStrings() []string {
	s := make([]string, len(g))
	for i, node := range g {
		s[i] = node.name
	}
	return s
}

// Displays the dependency graph
func displayGraph(graph Graph) {
	for _, node := range graph {
		for _, dep := range node.deps {
			fmt.Printf("%s -> %s\n", node.name, dep.name)
		}
	}
}
