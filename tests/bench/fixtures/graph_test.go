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
	nodeA := NewNode("A")
	nodeB := NewNode("B")
	nodeC := NewNode("C", "A")
	nodeD := NewNode("D", "B")
	nodeE := NewNode("E", "C", "D")
	nodeF := NewNode("F", "A", "B")
	nodeG := NewNode("G", "E", "F")
	nodeH := NewNode("H", "G")
	nodeI := NewNode("I", "A")
	nodeJ := NewNode("J", "B")
	nodeK := NewNode("K")

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
	nodeA := NewNode("A", "I")
	nodeB := NewNode("B")
	nodeC := NewNode("C", "A")
	nodeD := NewNode("D", "B")
	nodeE := NewNode("E", "C", "D")
	nodeF := NewNode("F", "A", "B")
	nodeG := NewNode("G", "E", "F")
	nodeH := NewNode("H", "G")
	nodeI := NewNode("I", "A") // <-- circular
	nodeJ := NewNode("J", "B")
	nodeK := NewNode("K")

	var brokenGraph Graph
	brokenGraph = append(brokenGraph, nodeA, nodeB, nodeC, nodeD, nodeE, nodeF, nodeG, nodeH, nodeI, nodeJ, nodeK)

	_, err := resolveGraph(brokenGraph)
	require.Error(t, err)

}
