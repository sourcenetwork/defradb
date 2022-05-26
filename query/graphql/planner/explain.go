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

import (
	"github.com/iancoleman/strcase"
)

type explainablePlanNode interface {
	planNode
	Explain() (map[string]interface{}, error)
}

// Compile time check for all planNodes that should be explainable (satisfy explainablePlanNode).
var (
	_ explainablePlanNode = (*createNode)(nil)
	_ explainablePlanNode = (*deleteNode)(nil)
	_ explainablePlanNode = (*scanNode)(nil)
	_ explainablePlanNode = (*selectNode)(nil)
	_ explainablePlanNode = (*selectTopNode)(nil)
	_ explainablePlanNode = (*typeIndexJoin)(nil)

	// Nodes to implement in the next explain request PRs.
	// _ explainablePlanNode = (*averageNode)(nil)
	// _ explainablePlanNode = (*commitSelectNode)(nil)
	// _ explainablePlanNode = (*countNode)(nil)
	// _ explainablePlanNode = (*dagScanNode)(nil)
	// _ explainablePlanNode = (*renderNode)(nil)
	// _ explainablePlanNode = (*sortNode)(nil)
	// _ explainablePlanNode = (*sumNode)(nil)
	// _ explainablePlanNode = (*updateNode)(nil)

	// Internal Nodes that we don't want to expose / explain.
	// - commitSelectTopNode
	// - renderLimitNode
	// - groupNode
	// - hardLimitNode
	// - headsetScanNode
	// - parallelNode
	// - pipeNode
	// - typeJoinMany
	// - typeJoinOne
)

// buildExplainGraph builds the explainGraph from the given top level plan.
//
// Request:
// query @explain {
//     user {
//       _key
//       age
//       name
//     }
// }
//
//  Response:
// {
//   "data": [
//     {
//       "explain": {
//         "selectTopNode": {
//           "selectNode": {
//             "filter": null,
//             "scanNode": {
//               "collectionID": "1",
//               "collectionName": "user",
//               "filter": null
//             }
//           }
//         }
//       }
//     }
//   ]
// }
func buildExplainGraph(source planNode) (map[string]interface{}, error) {

	explainGraph := map[string]interface{}{}

	if source == nil {
		return explainGraph, nil
	}

	switch node := source.(type) {

	// Walk the multiple children if it is a MultiNode.
	// Note: MultiNode nodes are not explainable but we use them to wrap the children under them.
	case MultiNode:
		// List to store all explain graphs of explainable children of MultiNode.
		multiChildExplainGraph := []map[string]interface{}{}
		for _, childSource := range node.Children() {
			childExplainGraph, err := buildExplainGraph(childSource)
			if err != nil {
				return nil, err
			}
			// Add the child's explain graph to the list with all explainable children of MultiNode.
			multiChildExplainGraph = append(multiChildExplainGraph, childExplainGraph)
		}
		// Add the list of explainable children graphs we built above under the current MultiNode.
		explainNodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[explainNodeLabelTitle] = multiChildExplainGraph

	// If this node has subscribed to the optable-interface that makes a node explainable.
	case explainablePlanNode:
		// Start building the explain graph.
		explainGraphBuilder, err := node.Explain()
		if err != nil {
			return nil, err
		}
		// If not the last child then keep walking the graph to find more explainable nodes.
		if node.Source() != nil {
			nextExplainGraph, err := buildExplainGraph(node.Source())
			if err != nil {
				return nil, err
			}
			// Add the key-value pairs from the next nested explain graph into the builder.
			for key, value := range nextExplainGraph {
				explainGraphBuilder[key] = value
			}
		}
		// Add the explainable graph of the next node under current explainable node.
		explainNodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[explainNodeLabelTitle] = explainGraphBuilder

	default:
		// Node is neither a MultiNode nor an "explainable" node. Skip over it but walk it's child(ren).
		var err error
		explainGraph, err = buildExplainGraph(source.Source())
		if err != nil {
			return nil, err
		}
	}

	return explainGraph, nil

}
