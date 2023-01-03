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
	Explain() (map[string]any, error)
}

// Compile time check for all planNodes that should be explainable (satisfy explainablePlanNode).
var (
	_ explainablePlanNode = (*averageNode)(nil)
	_ explainablePlanNode = (*countNode)(nil)
	_ explainablePlanNode = (*createNode)(nil)
	_ explainablePlanNode = (*dagScanNode)(nil)
	_ explainablePlanNode = (*deleteNode)(nil)
	_ explainablePlanNode = (*groupNode)(nil)
	_ explainablePlanNode = (*limitNode)(nil)
	_ explainablePlanNode = (*orderNode)(nil)
	_ explainablePlanNode = (*scanNode)(nil)
	_ explainablePlanNode = (*selectNode)(nil)
	_ explainablePlanNode = (*selectTopNode)(nil)
	_ explainablePlanNode = (*sumNode)(nil)
	_ explainablePlanNode = (*topLevelNode)(nil)
	_ explainablePlanNode = (*typeIndexJoin)(nil)
	_ explainablePlanNode = (*updateNode)(nil)
)

const (
	childFieldNameLabel = "childFieldName"
	collectionIDLabel   = "collectionID"
	collectionNameLabel = "collectionName"
	dataLabel           = "data"
	fieldNameLabel      = "fieldName"
	filterLabel         = "filter"
	idsLabel            = "ids"
	limitLabel          = "limit"
	offsetLabel         = "offset"
	sourcesLabel        = "sources"
	spansLabel          = "spans"
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
//	           ...
//             "scanNode": {
//               ...
//             }
//           }
//         }
//       }
//     }
//   ]
// }
func buildExplainGraph(source planNode) (map[string]any, error) {
	explainGraph := map[string]any{}

	if source == nil {
		return explainGraph, nil
	}

	switch node := source.(type) {
	// Walk the multiple children if it is a MultiNode.
	// Note: MultiNode nodes are not explainable but we use them to wrap the children under them.
	case MultiNode:
		// List to store all explain graphs of explainable children of MultiNode.
		multiChildExplainGraph := []map[string]any{}
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

	// For typeIndexJoin restructure the graphs to show both `root` and `subType` at the same level.
	case *typeIndexJoin:
		// Get the non-restructured explain graph.
		indexJoinGraph, err := node.Explain()
		if err != nil {
			return nil, err
		}

		// If not the last child then keep walking and explaining the root graph,
		// as long as there are more explainable nodes left under root.
		if node.Source() != nil {
			indexJoinRootExplainGraph, err := buildExplainGraph(node.Source())
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "root" graph.
			indexJoinGraph["root"] = indexJoinRootExplainGraph
		}
		// Add this restructured typeIndexJoin explain graph.
		explainGraph[strcase.ToLowerCamel(node.Kind())] = indexJoinGraph

	// If this node has subscribed to the optable-interface that makes a node explainable.
	case explainablePlanNode:
		// Start building the explain graph.
		explainGraphBuilder, err := node.Explain()
		if err != nil {
			return nil, err
		}

		// Support nil to signal as if there are no attributes to explain for that node.
		if explainGraphBuilder == nil {
			explainGraphBuilder = map[string]any{}
		}

		// If not the last child then keep walking the graph to find more explainable nodes.
		// Also make sure the next source / child isn't a recursive `topLevelNode`.
		if next := node.Source(); next != nil && next.Kind() != topLevelNodeKind {
			nextExplainGraph, err := buildExplainGraph(next)
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
