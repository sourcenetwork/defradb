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
	"encoding/json"

	"github.com/iancoleman/strcase"
	plannerTypes "github.com/sourcenetwork/defradb/query/graphql/planner/types"
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

	// Nodes to implement in the next explain request PRs.
	// _ explainablePlanNode = (*averageNode)(nil)
	// _ explainablePlanNode = (*commitSelectNode)(nil)
	// _ explainablePlanNode = (*countNode)(nil)
	// _ explainablePlanNode = (*dagScanNode)(nil)
	// _ explainablePlanNode = (*renderNode)(nil)
	// _ explainablePlanNode = (*sortNode)(nil)
	// _ explainablePlanNode = (*sumNode)(nil)
	// _ explainablePlanNode = (*typeIndexJoin)(nil)
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

	// fmt.Println("======================= source      : ", source.Kind())

	explainGraph := map[string]interface{}{}

	if source == nil {
		return explainGraph, nil
	}

	switch node := source.(type) {

	// Walk the multiple children if it is a MultiNode (MultiNode itself is non-explainable).
	case MultiNode:
		childrenSources := node.Children()
		for _, childSource := range childrenSources {
			var err error
			explainGraph, err = buildExplainGraph(childSource.Source())
			if err != nil {
				return nil, err
			}
		}

	// Only explain the node if it is explainable.
	case explainablePlanNode:
		explainGraphBuilder, err := node.Explain()
		if err != nil {
			return nil, err
		}

		// If not the last child then keep walking the graph to find more explainable nodes.
		if node.Source() != nil {
			childExplainGraph, err := buildExplainGraph(node.Source())
			if err != nil {
				return nil, err
			}
			for key, value := range childExplainGraph {
				explainGraphBuilder[key] = value
			}
		}

		explainNodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[explainNodeLabelTitle] = explainGraphBuilder

	default:
		// Node is neither a MultiNode or any explainable node. Skip over it but walk it's children.
		var err error
		explainGraph, err = buildExplainGraph(source.Source())
		if err != nil {
			return nil, err
		}
	}

	return explainGraph, nil

}

// Following are all the planNodes that are subscribing to the explainablePlanNode.

func (n *selectTopNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{
		// No attributes are returned for selectTopNode.
	}

	return explainerMap, nil
}

func (n *selectNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{}

	// Add the filter attribute if it exists.
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[plannerTypes.Filter] = nil
	} else {
		explainerMap[plannerTypes.Filter] = n.filter.Conditions
	}

	return explainerMap, nil
}

func (n *scanNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{}

	// Add the filter attribute if it exists.
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[plannerTypes.Filter] = nil
	} else {
		explainerMap[plannerTypes.Filter] = n.filter.Conditions
	}

	// Add the collection attributes.
	explainerMap[plannerTypes.CollectionName] = n.desc.Name
	explainerMap[plannerTypes.CollectionID] = n.desc.IDString()

	// @todo: Add the index attribute.

	// @todo: Add the spans attribute (couldn't find an example to test).
	// spansAttribute := styleAttribute("Spans")
	// explainerMap[spansAttribute] = n.spans

	return explainerMap, nil
}

func (n *createNode) Explain() (map[string]interface{}, error) {

	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(n.newDocStr), &data)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		plannerTypes.Data: data,
	}, nil
}

func (n *deleteNode) Explain() (map[string]interface{}, error) {

	explainerMap := map[string]interface{}{}

	// Add the document id(s) that request wants to delete.
	explainerMap[plannerTypes.IDs] = n.ids

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[plannerTypes.Filter] = nil
	} else {
		explainerMap[plannerTypes.Filter] = n.filter.Conditions
	}

	return explainerMap, nil
}
