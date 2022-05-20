// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

// Request:
//     query @explain {
//         user {
//             _key
//         }
//     }

// Response:
//{
//  "data": [
//    {
//      "explain": {
//        "Node => selectTopNode": {
//          "Attribite": "Select Top Node",
//          "Node => selectNode": {
//            "Attribite": "Select Node",
//            "Node => scanNode": {
//              "Attribite": "Scan Node"
//            }
//          }
//        }
//      }
//    }
//  ]
//}

func buildExplainGraph(source planNode) map[string]interface{} {

	var explainGraph map[string]interface{} = map[string]interface{}{}

	if source == nil {
		return explainGraph
	}

	// Walk the multiple children if it is a MultiNode (MultiNode itself is non-explainable).
	multiNode, isMultiNode := source.(MultiNode)
	if isMultiNode {
		childrenSources := multiNode.Children()
		for _, childSource := range childrenSources {
			explainGraph = buildExplainGraph(childSource.Source())
		}
	}

	// Only explain the node if it is explainable.
	explainableSource, isExplainable := source.(explainablePlanNode)
	if isExplainable {
		explainGraphBuilder := explainableSource.Explain()

		// If not the last child then keep walking the graph to find more explainable nodes.
		notLeafSource := explainableSource.Source() != nil
		if notLeafSource {
			childExplainGraph := buildExplainGraph(explainableSource.Source())
			for key, value := range childExplainGraph {
				explainGraphBuilder[key] = value
			}
		}

		explainNodeLabelTitle := "Node => " + explainableSource.Kind()
		explainGraph[explainNodeLabelTitle] = explainGraphBuilder
	}

	return explainGraph

}

type explainablePlanNode interface {
	planNode
	Explain() map[string]interface{}
}

// Compile time check for all planNodes that should be explainable (satisfy explainablePlanNode).
var (
	// _ explainablePlanNode = (*averageNode)(nil)
	// _ explainablePlanNode = (*commitSelectNode)(nil)
	// _ explainablePlanNode = (*countNode)(nil)
	// _ explainablePlanNode = (*createNode)(nil)
	// _ explainablePlanNode = (*dagScanNode)(nil)
	// _ explainablePlanNode = (*deleteNode)(nil)
	// _ explainablePlanNode = (*renderNode)(nil)
	_ explainablePlanNode = (*scanNode)(nil)
	_ explainablePlanNode = (*selectNode)(nil)
	_ explainablePlanNode = (*selectTopNode)(nil)
	// _ explainablePlanNode = (*sortNode)(nil)
	// _ explainablePlanNode = (*sumNode)(nil)
	// _ explainablePlanNode = (*typeIndexJoin)(nil)
	// _ explainablePlanNode = (*updateNode)(nil)

	// Internal Nodes that we don't want to expose / explain.

	// _ explainablePlanNode = (*commitSelectTopNode)(nil)
	// _ explainablePlanNode = (*renderLimitNode)(nil)
	// _ explainablePlanNode = (*groupNode)(nil)
	// _ explainablePlanNode = (*hardLimitNode)(nil)
	// _ explainablePlanNode = (*headsetScanNode)(nil)
	// _ explainablePlanNode = (*parallelNode)(nil)
	// _ explainablePlanNode = (*pipeNode)(nil)
	// _ explainablePlanNode = (*typeJoinMany)(nil)
	// _ explainablePlanNode = (*typeJoinOne)(nil)
)

// Following are all the planNodes that are subscribing to the explainablePlanNode.

func (n *selectTopNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{
		"Attribite": "Select Top Node",
	}

	return explainerMap
}

func (n *selectNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{
		"Attribite": "Select Node",
	}
	return explainerMap
}

func (n *scanNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{
		"Attribite": "Scan Node",
	}

	return explainerMap
}
