// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

/*
A MultiNode is a planNode which contains multiple sub nodes,
that can be executed either in parallel, and serial. Each Values()
response is added to the stored document. Each child node is a named
planNode, where the name is the target field for the planNode.

This is also the basis of the MultiScannerNode. The MultiScannerNode
is a MultiNode, which shares an underlying scanNode. Each step of a
MultiScannerNode takes one value from the source node, and uses its
results in all the attached multinodes.
*/

type MultiNode interface {
	planNode
	Children() []planNode
}

// parallelNode implements the MultiNode interface. It
// enables parallel execution of planNodes. This is needed
// if a single request has multiple Select statements at the
// same depth in the request.
// Eg:
//
//	user {
//			_docID
//			name
//			friends {
//				name
//			}
//			_version {
//				cid
//			}
//	}
//
// In this example, both the friends selection and the _version
// selection require their own planNode sub graphs to complete.
// However, they are entirely independent graphs, so they can
// be executed in parallel.
type parallelNode struct { // serialNode?
	documentIterator
	docMapper

	p *Planner

	children     []planNode
	childIndexes []int

	source    planNode
	multiscan *multiScanNode
}

func (p *parallelNode) applyToPlans(fn func(n planNode) error) error {
	for _, plan := range p.children {
		if err := fn(plan); err != nil {
			return err
		}
	}
	return nil
}

func (p *parallelNode) Kind() string {
	return "parallelNode"
}

func (p *parallelNode) Init() error {
	newChildren := make([]planNode, len(p.children))
	newChildIndexes := make([]int, len(p.childIndexes))

	endIndex := len(p.children) - 1
	startIndex := 0
	for i, child := range p.children {
		switch child.(type) {
		case *dagScanNode:
			// Any node types that result in `nextAppend` calls must be executed last, as they are
			// dependent on the docID set by the `nextMerge` calls.
			//
			// For example, there might be two children, a `scanNode` and a `dagScanNode`. As per the logic
			// in `parallelNode.Next`, the `scanNode` must be executed before the `dagScanNode` as the
			// `dagScanNode` uses the DocID yielded from the `scanNode`.
			newChildren[endIndex] = child
			newChildIndexes[endIndex] = p.childIndexes[i]
			endIndex--
		default:
			newChildren[startIndex] = child
			newChildIndexes[startIndex] = p.childIndexes[i]
			startIndex++
		}
	}

	p.children = newChildren
	p.childIndexes = newChildIndexes

	return p.applyToPlans(func(n planNode) error {
		return n.Init()
	})
}

func (p *parallelNode) Start() error {
	return p.applyToPlans(func(n planNode) error {
		return n.Start()
	})
}

func (p *parallelNode) Prefixes(prefixes []keys.Walkable) {
	_ = p.applyToPlans(func(n planNode) error {
		n.Prefixes(prefixes)
		return nil
	})
}

func (p *parallelNode) Close() error {
	return p.applyToPlans(func(n planNode) error {
		return n.Close()
	})
}

// Next loops through all the children nodes, and calls Next().
// It only needs a single child plan to return true for it
// to return true. Same with errors.
func (p *parallelNode) Next() (bool, error) {
	p.currentValue = p.documentMapping.NewDoc()

	var orNext bool
	for i, plan := range p.children {
		var next bool
		var err error
		switch n := plan.(type) {
		case *dagScanNode:
			next, err = p.nextAppend(i, n)
		default: // anything else is a merge
			next, err = p.nextMerge(i, n)
		}
		if err != nil {
			return false, err
		}
		orNext = orNext || next
	}
	// if none of the children return true for next, then this will be false.
	// if ANY of the children return true, this will be true (logical OR)
	return orNext, nil
}

func (p *parallelNode) nextMerge(_ int, plan planNode) (bool, error) {
	if next, err := plan.Next(); !next {
		return false, err
	}

	// Field-by-fields check is necessary because parallelNode can have multiple children, and
	// each child can return the same doc, but with different related fields available
	// depending on what is requested.
	newFields := plan.Value().Fields
	for i := range newFields {
		if p.currentValue.Fields[i] == nil {
			p.currentValue.Fields[i] = newFields[i]
		}
	}

	return true, nil
}

func (p *parallelNode) nextAppend(index int, plan planNode) (bool, error) {
	key := p.currentValue.GetID()
	if key == "" {
		return false, nil
	}

	// pass the doc key as a reference through the prefixes interface
	prefixes := []keys.Walkable{keys.DataStoreKey{DocID: key}}
	plan.Prefixes(prefixes)
	err := plan.Init()
	if err != nil {
		return false, err
	}

	results := make([]core.Doc, 0)
	for {
		next, err := plan.Next()
		if err != nil {
			return false, err
		}

		if !next {
			break
		}

		results = append(results, plan.Value())
	}
	p.currentValue.Fields[p.childIndexes[index]] = results
	return true, nil
}

func (p *parallelNode) Source() planNode { return p.source }

func (p *parallelNode) Children() []planNode {
	return p.children
}

func (p *parallelNode) addChild(fieldIndex int, node planNode) {
	p.children = append(p.children, node)
	p.childIndexes = append(p.childIndexes, fieldIndex)
}

func (n *selectNode) addSubPlan(fieldIndex int, newPlan planNode) error {
	switch sourceNode := n.source.(type) {
	// if its a scan node, we either replace or create a multinode
	case *updateNode, *scanNode, *pipeNode:
		switch newPlan.(type) {
		case *typeIndexJoin:
			n.source = newPlan
		case *dagScanNode:
			m := &parallelNode{
				p:         n.planner,
				source:    newPlan,
				docMapper: docMapper{n.source.DocumentMap()},
			}
			m.addChild(-1, n.source)
			m.addChild(fieldIndex, newPlan)
			n.source = m
		default:
			return client.NewErrUnhandledType("sub plan", newPlan)
		}

	case *typeIndexJoin:
		var multiscan *multiScanNode
		var source planNode
		var origSource planNode

		// we need to replace the original "source" with the appropriate
		// multiScanNode type. However, not all "source" types are equal.
		// For query ops the target source is a `*scanNode`, for mutations
		// the target source is either a `*createNode`, `*updateNode`, or
		// a `*upsertNode`
		//
		// This is necessary since the `*typeIndexJoin` join will read
		// from this source multiple times per iteration, so we need
		// to make sure that we're caching the necessary state. Eg during
		// a mutation multiple reads shouldn't trigger multiple mutations.
		switch n.origSource.(type) {
		case *updateNode:
			origSource, _ = walkAndFindPlanType[*updateNode](newPlan)
		default:
			origSource, _ = walkAndFindPlanType[*scanNode](newPlan)
		}

		if origSource != nil {
			multiscan = &multiScanNode{planNode: origSource}
			if err := n.planner.walkAndReplacePlan(n.source, origSource, multiscan); err != nil {
				return err
			}
			if err := n.planner.walkAndReplacePlan(newPlan, origSource, multiscan); err != nil {
				return err
			}
			multiscan.addReader()
			multiscan.addReader()

			source = multiscan
		} else {
			source = newPlan
		}

		parallelNode := &parallelNode{
			p:         n.planner,
			multiscan: multiscan,
			source:    source,
			docMapper: docMapper{n.source.DocumentMap()},
		}
		parallelNode.addChild(-1, n.source)
		parallelNode.addChild(fieldIndex, newPlan)
		n.source = parallelNode

	// we already have an existing parallelNode as our source
	case *parallelNode:
		switch newPlan.(type) {
		// We have a internal multiscanNode on our MultiNode
		case *scanNode, *typeIndexJoin:
			if sourceNode.multiscan != nil {
				if err := n.planner.walkAndReplacePlan(newPlan, sourceNode.multiscan.Source(), sourceNode.multiscan); err != nil {
					return err
				}
				sourceNode.multiscan.addReader()
			}
		}

		if _, ok := newPlan.(*typeIndexJoin); ok {
			for i, child := range sourceNode.children {
				if _, ok := child.(*scanNode); ok {
					// `typeIndexJoin` contain the `scanNode`, so if the `scanNode` has already been added to the
					// `parallelNode` it must be overwritten by the `typeIndexJoin` that wraps it.
					sourceNode.children[i] = newPlan
					sourceNode.childIndexes[i] = fieldIndex
					return nil
				}
			}
		}

		sourceNode.addChild(fieldIndex, newPlan)
	}

	return nil
}
