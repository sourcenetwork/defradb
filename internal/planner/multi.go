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

func (n *parallelNode) Kind() string {
	return "parallelNode"
}

func (p *parallelNode) Init() error {
	return p.applyToPlans(func(n planNode) error {
		return n.Init()
	})
}

func (p *parallelNode) Start() error {
	return p.applyToPlans(func(n planNode) error {
		return n.Start()
	})
}

func (p *parallelNode) Spans(spans core.Spans) {
	_ = p.applyToPlans(func(n planNode) error {
		n.Spans(spans)
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
		// isMerge := false
		switch n := plan.(type) {
		case *scanNode, *typeIndexJoin:
			// isMerge = true
			next, err = p.nextMerge(i, n)
		case *dagScanNode:
			next, err = p.nextAppend(i, n)
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

	doc := plan.Value()
	copy(p.currentValue.Fields, doc.Fields)

	return true, nil
}

func (p *parallelNode) nextAppend(index int, plan planNode) (bool, error) {
	key := p.currentValue.GetID()
	if key == "" {
		return false, nil
	}

	// pass the doc key as a reference through the spans interface
	spans := core.NewSpans(core.NewSpan(core.DataStoreKey{DocID: key}, core.DataStoreKey{}))
	plan.Spans(spans)
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

func (p *parallelNode) Source() planNode { return p.multiscan }

func (p *parallelNode) Children() []planNode {
	return p.children
}

func (p *parallelNode) addChild(fieldIndex int, node planNode) {
	p.children = append(p.children, node)
	p.childIndexes = append(p.childIndexes, fieldIndex)
}

func (s *selectNode) addSubPlan(fieldIndex int, newPlan planNode) error {
	switch sourceNode := s.source.(type) {
	// if its a scan node, we either replace or create a multinode
	case *scanNode, *pipeNode:
		switch newPlan.(type) {
		case *scanNode, *typeIndexJoin:
			s.source = newPlan
		case *dagScanNode:
			m := &parallelNode{
				p:         s.planner,
				docMapper: docMapper{s.source.DocumentMap()},
			}
			m.addChild(-1, s.source)
			m.addChild(fieldIndex, newPlan)
			s.source = m
		default:
			return client.NewErrUnhandledType("sub plan", newPlan)
		}

	case *typeIndexJoin:
		origScan, _ := walkAndFindPlanType[*scanNode](newPlan)
		if origScan == nil {
			return ErrFailedToFindScanNode
		}
		// create our new multiscanner
		multiscan := &multiScanNode{scanNode: origScan}
		// replace our current source internal scanNode with our new multiscanner
		if err := s.planner.walkAndReplacePlan(s.source, origScan, multiscan); err != nil {
			return err
		}
		// create parallelNode
		parallelNode := &parallelNode{
			p:         s.planner,
			multiscan: multiscan,
			docMapper: docMapper{s.source.DocumentMap()},
		}
		parallelNode.addChild(-1, s.source)
		multiscan.addReader()
		// replace our new node internal scanNode with our new multiscanner
		if err := s.planner.walkAndReplacePlan(newPlan, origScan, multiscan); err != nil {
			return err
		}
		// add our newly updated plan to the multinode
		parallelNode.addChild(fieldIndex, newPlan)
		multiscan.addReader()
		s.source = parallelNode

	// we already have an existing parallelNode as our source
	case *parallelNode:
		switch newPlan.(type) {
		// We have a internal multiscanNode on our MultiNode
		case *scanNode, *typeIndexJoin:
			// replace our new node internal scanNode with our existing multiscanner
			if err := s.planner.walkAndReplacePlan(newPlan, sourceNode.multiscan.Source(), sourceNode.multiscan); err != nil {
				return err
			}
			sourceNode.multiscan.addReader()
		}

		sourceNode.addChild(fieldIndex, newPlan)
	}
	return nil
}
