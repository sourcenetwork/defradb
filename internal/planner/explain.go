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
	"context"
	"strconv"

	"github.com/iancoleman/strcase"

	"github.com/sourcenetwork/defradb/client/request"
)

type explainablePlanNode interface {
	planNode

	// Explain returns explain datapoints that are scoped to this node.
	//
	// It is possible that no datapoint is gathered for a certain node.
	//
	// Explain with type execute should NOT be called before the `Next()` has been called.
	Explain(explainType request.ExplainType) (map[string]any, error)
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
	inputLabel          = "input"
	fieldNameLabel      = "fieldName"
	filterLabel         = "filter"
	joinRootLabel       = "root"
	joinSubTypeLabel    = "subType"
	limitLabel          = "limit"
	offsetLabel         = "offset"
	sourcesLabel        = "sources"
	spansLabel          = "spans"
)

// buildDebugExplainGraph dumps the entire plan graph as is, with all the plan nodes.
//
// Note: This also includes plan nodes that aren't "explainable".
func buildDebugExplainGraph(source planNode) (map[string]any, error) {
	explainGraph := map[string]any{}

	if source == nil {
		return explainGraph, nil
	}

	switch node := source.(type) {
	// Walk the multiple children if it is a MultiNode.
	case MultiNode:
		multiChildExplainGraph := []map[string]any{}
		for _, childSource := range node.Children() {
			childExplainGraph, err := buildDebugExplainGraph(childSource)
			if err != nil {
				return nil, err
			}
			multiChildExplainGraph = append(multiChildExplainGraph, childExplainGraph)
		}
		nodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[nodeLabelTitle] = multiChildExplainGraph

	case *typeJoinMany:
		var explainGraphBuilder = map[string]any{}

		// If root is not the last child then keep walking and explaining the root graph.
		if node.parentPlan != nil {
			indexJoinRootExplainGraph, err := buildDebugExplainGraph(node.parentPlan)
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "root" graph.
			explainGraphBuilder[joinRootLabel] = indexJoinRootExplainGraph
		}

		if node.childPlan != nil {
			indexJoinSubTypeExplainGraph, err := buildDebugExplainGraph(node.childPlan)
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "subType" graph.
			explainGraphBuilder[joinSubTypeLabel] = indexJoinSubTypeExplainGraph
		}

		nodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[nodeLabelTitle] = explainGraphBuilder

	case *typeJoinOne:
		var explainGraphBuilder = map[string]any{}

		// If root is not the last child then keep walking and explaining the root graph.
		if node.parentPlan != nil {
			indexJoinRootExplainGraph, err := buildDebugExplainGraph(node.parentPlan)
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "root" graph.
			explainGraphBuilder[joinRootLabel] = indexJoinRootExplainGraph
		} else {
			explainGraphBuilder[joinRootLabel] = nil
		}

		if node.childPlan != nil {
			indexJoinSubTypeExplainGraph, err := buildDebugExplainGraph(node.childPlan)
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "subType" graph.
			explainGraphBuilder[joinSubTypeLabel] = indexJoinSubTypeExplainGraph
		} else {
			explainGraphBuilder[joinSubTypeLabel] = nil
		}

		nodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[nodeLabelTitle] = explainGraphBuilder

	default:
		var explainGraphBuilder = map[string]any{}

		// If not the last child then keep walking the graph to find more plan nodes.
		// Also make sure the next source / child isn't a recursive `topLevelNode`.
		if next := node.Source(); next != nil && next.Kind() != topLevelNodeKind {
			var err error
			explainGraphBuilder, err = buildDebugExplainGraph(next)
			if err != nil {
				return nil, err
			}
		}
		// Add the graph of the next node under current node.
		nodeLabelTitle := strcase.ToLowerCamel(node.Kind())
		explainGraph[nodeLabelTitle] = explainGraphBuilder
	}

	return explainGraph, nil
}

// buildSimpleExplainGraph builds the explainGraph from the given top level plan.
//
// Request:
//
//	query @explain {
//	    user {
//	      _docID
//	      age
//	      name
//	    }
//	}
//
// Response:
//
//	{
//	  "data": [
//	    {
//	      "explain": {
//	        "selectTopNode": {
//	          "selectNode": {
//		           ...
//	            "scanNode": {
//	              ...
//	            }
//	          }
//	        }
//	      }
//	    }
//	  ]
//	}
func buildSimpleExplainGraph(source planNode) (map[string]any, error) {
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
			childExplainGraph, err := buildSimpleExplainGraph(childSource)
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
		indexJoinGraph, err := node.Explain(request.SimpleExplain)
		if err != nil {
			return nil, err
		}

		// If not the last child then keep walking and explaining the root graph,
		// as long as there are more explainable nodes left under root.
		if node.Source() != nil {
			indexJoinRootExplainGraph, err := buildSimpleExplainGraph(node.Source())
			if err != nil {
				return nil, err
			}
			// Add the explaination of the rest of the explain graph under the "root" graph.
			indexJoinGraph[joinRootLabel] = indexJoinRootExplainGraph
		}
		// Add this restructured typeIndexJoin explain graph.
		explainGraph[strcase.ToLowerCamel(node.Kind())] = indexJoinGraph

	// If this node has subscribed to the optable-interface that makes a node explainable.
	case explainablePlanNode:
		// Start building the explain graph.
		explainGraphBuilder, err := node.Explain(request.SimpleExplain)
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
			nextExplainGraph, err := buildSimpleExplainGraph(next)
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
		explainGraph, err = buildSimpleExplainGraph(source.Source())
		if err != nil {
			return nil, err
		}
	}

	return explainGraph, nil
}

// collectExecuteExplainInfo structures and returns the already collected information
// when the request was executed with the explain option.
//
// Note: Can only be called once the entire plan has been executed.
func collectExecuteExplainInfo(executedPlan planNode) (map[string]any, error) {
	executeExplainInfo := map[string]any{}

	if executedPlan == nil {
		return executeExplainInfo, nil
	}

	switch executedNode := executedPlan.(type) {
	case MultiNode:
		multiChildExplainGraph := []map[string]any{}
		for _, childSource := range executedNode.Children() {
			childExplainGraph, err := collectExecuteExplainInfo(childSource)
			if err != nil {
				return nil, err
			}
			multiChildExplainGraph = append(multiChildExplainGraph, childExplainGraph)
		}
		explainNodeLabelTitle := strcase.ToLowerCamel(executedNode.Kind())
		executeExplainInfo[explainNodeLabelTitle] = multiChildExplainGraph

	case explainablePlanNode:
		executeExplainBuilder, err := executedNode.Explain(request.ExecuteExplain)
		if err != nil {
			return nil, err
		}

		if executeExplainBuilder == nil {
			executeExplainBuilder = map[string]any{}
		}

		if next := executedNode.Source(); next != nil && next.Kind() != topLevelNodeKind {
			nextExplainGraph, err := collectExecuteExplainInfo(next)
			if err != nil {
				return nil, err
			}
			for key, value := range nextExplainGraph {
				executeExplainBuilder[key] = value
			}
		}
		explainNodeLabelTitle := strcase.ToLowerCamel(executedNode.Kind())
		executeExplainInfo[explainNodeLabelTitle] = executeExplainBuilder

	default:
		var err error
		executeExplainInfo, err = collectExecuteExplainInfo(executedPlan.Source())
		if err != nil {
			return nil, err
		}
	}

	return executeExplainInfo, nil
}

// executeAndExplainRequest executes the plan graph gathering the information/datapoints
// during the execution. Then once the execution is complete returns the collected datapoints.
//
// Note: This function only fails if the collection of the datapoints goes wrong, otherwise
// even if plan execution fails this function would return the collected datapoints.
func (p *Planner) executeAndExplainRequest(
	ctx context.Context,
	plan planNode,
) ([]map[string]any, error) {
	executionSuccess := false
	planExecutions := uint64(0)

	if err := plan.Start(); err != nil {
		return []map[string]any{
			{
				request.ExplainLabel: map[string]any{
					"executionSuccess": executionSuccess,
					"executionErrors":  []string{"plan failed to start"},
					"planExecutions":   planExecutions,
				},
			},
		}, nil
	}

	next, err := plan.Next()
	planExecutions++
	if err != nil {
		return []map[string]any{
			{
				request.ExplainLabel: map[string]any{
					"executionSuccess": executionSuccess,
					"executionErrors": []string{
						"failure at plan execution count: " + strconv.FormatUint(planExecutions, 10),
						err.Error(),
					},
					"planExecutions": planExecutions,
				},
			},
		}, nil
	}

	docs := []map[string]any{}
	docMap := plan.DocumentMap()

	for next {
		copy := docMap.ToMap(plan.Value())
		docs = append(docs, copy)

		next, err = plan.Next()
		planExecutions++

		if err != nil {
			return []map[string]any{
				{
					request.ExplainLabel: map[string]any{
						"executionSuccess": executionSuccess,
						"executionErrors": []string{
							"failure at plan execution count: " + strconv.FormatUint(planExecutions, 10),
							err.Error(),
						},
						"planExecutions":    planExecutions,
						"sizeOfResultSoFar": len(docs),
					},
				},
			}, nil
		}
	}
	executionSuccess = true

	executeExplain, err := collectExecuteExplainInfo(plan)
	if err != nil {
		return nil, NewErrFailedToCollectExecExplainInfo(err)
	}

	executeExplain["executionSuccess"] = executionSuccess
	executeExplain["planExecutions"] = planExecutions
	executeExplain["sizeOfResult"] = len(docs)

	return []map[string]any{
		{
			request.ExplainLabel: executeExplain,
		},
	}, err
}

// explainRequest explains the given request plan according to the type of explain request.
func (p *Planner) explainRequest(
	ctx context.Context,
	plan planNode,
	explainType request.ExplainType,
) ([]map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		// walks through the plan graph, and outputs the concrete planNodes that should
		// be executed, maintaining their order in the plan graph (does not actually execute them).
		explainGraph, err := buildSimpleExplainGraph(plan)
		if err != nil {
			return nil, err
		}

		explainResult := []map[string]any{
			{
				request.ExplainLabel: explainGraph,
			},
		}

		return explainResult, nil

	case request.DebugExplain:
		// walks through the plan graph, and outputs the concrete planNodes that should
		// be executed, maintaining their order in the plan graph (does not actually execute them).
		explainGraph, err := buildDebugExplainGraph(plan)
		if err != nil {
			return nil, err
		}

		explainResult := []map[string]any{
			{
				request.ExplainLabel: explainGraph,
			},
		}

		return explainResult, nil

	case request.ExecuteExplain:
		return p.executeAndExplainRequest(ctx, plan)

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
