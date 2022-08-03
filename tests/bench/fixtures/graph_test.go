package fixtures

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraph(t *testing.T) {
	//
	// A working dependency graph
	//
	nodeA := NewNode("A", nil)
	nodeB := NewNode("B", nil)
	nodeC := NewNode("C", nil, NewNode("A", nil))
	nodeD := NewNode("D", nil, NewNode("B", nil))
	nodeE := NewNode("E", nil, NewNode("C", nil), NewNode("D", nil))
	nodeF := NewNode("F", nil, NewNode("A", nil), NewNode("B", nil))
	nodeG := NewNode("G", nil, NewNode("E", nil), NewNode("F", nil))
	nodeH := NewNode("H", nil, NewNode("G", nil))
	nodeI := NewNode("I", nil, NewNode("A", nil))
	nodeJ := NewNode("J", nil, NewNode("B", nil))
	nodeK := NewNode("K", nil)

	var workingGraph Graph
	workingGraph = append(workingGraph, nodeA, nodeB, nodeC, nodeD, nodeE, nodeF, nodeG, nodeH, nodeI, nodeJ, nodeK)

	resolved, err := resolveGraph(workingGraph)
	require.NoError(t, err)

	for _, node := range resolved {
		fmt.Println(node.name)
	}

	displayGraph(resolved)

}

func TestBrokenGraph(t *testing.T) {

	//
	// A broken dependency graph with circular dependency
	//
	nodeA := NewNode("A", nil, NewNode("I", nil))
	nodeB := NewNode("B", nil)
	nodeC := NewNode("C", nil, NewNode("A", nil))
	nodeD := NewNode("D", nil, NewNode("B", nil))
	nodeE := NewNode("E", nil, NewNode("C", nil), NewNode("D", nil))
	nodeF := NewNode("F", nil, NewNode("A", nil), NewNode("B", nil))
	nodeG := NewNode("G", nil, NewNode("E", nil), NewNode("F", nil))
	nodeH := NewNode("H", nil, NewNode("G", nil))
	nodeI := NewNode("I", nil, NewNode("A", nil))
	nodeJ := NewNode("J", nil, NewNode("B", nil))
	nodeK := NewNode("K", nil)

	var brokenGraph Graph
	brokenGraph = append(brokenGraph, nodeA, nodeB, nodeC, nodeD, nodeE, nodeF, nodeG, nodeH, nodeI, nodeJ, nodeK)

	_, err := resolveGraph(brokenGraph)
	require.Error(t, err)

}
