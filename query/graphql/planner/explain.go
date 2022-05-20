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
//query @explain {
//  user {
//    _key
//    age
//    name
//  }
//}

// Response:
//{
//  "data": [
//    {
//      "explain": {
//        "Node => selectTopNode": {
//          "Node => selectNode": {
//            "-> Filter": null,
//            "Node => scanNode": {
//              "-> CollectionID": "1",
//              "-> CollectionName": "user",
//              "-> Filter": null
//            }
//          }
//        }
//      }
//    }
//  ]
//}
const explainNodeStyler string = "Node => "
const explainAttributeStyler string = "-> "

func styleNode(nodeName string) string {
	return explainNodeStyler + nodeName
}

func styleAttribute(attributeName string) string {
	return explainAttributeStyler + attributeName
}

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

		explainNodeLabelTitle := styleNode(explainableSource.Kind())
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
	_ explainablePlanNode = (*scanNode)(nil)
	_ explainablePlanNode = (*selectNode)(nil)
	_ explainablePlanNode = (*selectTopNode)(nil)

	// Nodes to implement in the next explain request PRs.
	// _ explainablePlanNode = (*averageNode)(nil)
	// _ explainablePlanNode = (*commitSelectNode)(nil)
	// _ explainablePlanNode = (*countNode)(nil)
	// _ explainablePlanNode = (*createNode)(nil)
	// _ explainablePlanNode = (*dagScanNode)(nil)
	// _ explainablePlanNode = (*deleteNode)(nil)
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

// Following are all the planNodes that are subscribing to the explainablePlanNode.

func (n *selectTopNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{
		// No attributes are returned for selectTopNode.
	}

	return explainerMap
}

func (n *selectNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{}

	if n == nil {
		return explainerMap
	}

	// Add the filter attribute if it exists.
	filterAttribute := styleAttribute("Filter")
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[filterAttribute] = nil
	} else {
		explainerMap[filterAttribute] = n.filter.Conditions
	}

	return explainerMap
}

func (n *scanNode) Explain() map[string]interface{} {
	explainerMap := map[string]interface{}{}

	if n == nil {
		return explainerMap
	}

	// Add the filter attribute if it exists.
	filterAttribute := styleAttribute("Filter")
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[filterAttribute] = nil
	} else {
		explainerMap[filterAttribute] = n.filter.Conditions
	}

	// Add the collection attributes.
	collectionNameAttribute := styleAttribute("CollectionName")
	explainerMap[collectionNameAttribute] = n.desc.Name

	collectionIDAttribute := styleAttribute("CollectionID")
	explainerMap[collectionIDAttribute] = n.desc.IDString()

	// @todo: Add the index attribute.

	// @todo: Add the spans attribute (couldn't find an example to test).
	// spansAttribute := styleAttribute("Spans")
	// explainerMap[spansAttribute] = n.spans

	return explainerMap
}
