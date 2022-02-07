// Copyright 2020 Source Inc.
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
	"errors"

	"log"

	"github.com/sourcenetwork/defradb/core"
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
	AddChild(string, planNode) error
	SetMultiScanner(*multiScanNode)
}

// mergeNode is a special interface for the MultiNode
// system. A mergeNode provides an entire document
// in its Values() func, with all the specific and
// necessary fields and subfields already merged
// into the doc
type mergeNode interface {
	planNode
	Merge() bool
}

// appendNode is a special interface for the MultiNode
// system.
type appendNode interface {
	planNode
	Append() bool
}

// parallelNode implements the MultiNode interface. It
// enables parallel execution of planNodes. This is needed
// if a single query has multiple Select statements at the
// same depth in the query.
// Eg:
// user {
//		_key
// 		name
// 		friends {
// 			name
// 		}
// 		_version {
// 			cid
// 		}
// }
//
// In this example, both the friends selection and the _version
// selection require their own planNode sub graphs to complete.
// However, they are entirely independent graphs, so they can
// be executed in parallel.
//
type parallelNode struct { // serialNode?
	p *Planner

	children    []planNode
	childFields []string

	multiscan *multiScanNode

	doc map[string]interface{}
}

func (p *parallelNode) applyToPlans(fn func(n planNode) error) error {
	for _, plan := range p.children {
		if err := fn(plan); err != nil {
			return err
		}
	}
	return nil
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
	err := p.applyToPlans(func(n planNode) error {
		n.Spans(spans)
		return nil
	})
	if err != nil {
		log.Print("applying spans to plans failed : ", err)
	}
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
	p.doc = make(map[string]interface{})
	var orNext bool
	for i, plan := range p.children {
		var next bool
		var err error
		// isMerge := false
		switch n := plan.(type) {
		case mergeNode:
			// isMerge = true
			next, err = p.nextMerge(i, n)
		case appendNode:
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

func (p *parallelNode) nextMerge(index int, plan mergeNode) (bool, error) {
	if next, err := plan.Next(); !next {
		return false, err
	}

	doc := plan.Values()
	for k, v := range doc {
		p.doc[k] = v
	}
	return true, nil
}

/*

scan node
=========
{
	_key: bae-ALICE,
	name: Alice,
	points: 124,
	verified: false
}

typeJoin node(merge)
=============
{
	friends: [
		{
			_key: bae-BOB,
			name: bob,
			points: 99.9,
			verified: true,
		}
	]
}

output
======

{
	_key: bae-ALICE,
	name: Alice,
	points: 124,
	verified: false,

	friends: [
		{
			_key: bae-BOB,
			name: bob,
			points: 99.9,
			verified: true,
		}
	]
}

*/

func (p *parallelNode) nextAppend(index int, plan appendNode) (bool, error) {
	if key, ok := p.doc["_key"].(string); ok {
		// pass the doc key as a reference through the spans interface
		spans := core.Spans{core.NewSpan(core.NewKey(key), core.Key{})}
		plan.Spans(spans)
		err := plan.Init()
		if err != nil {
			return false, err
		}
	} else {
		return false, nil
	}

	results := make([]map[string]interface{}, 0)
	for {
		next, err := plan.Next()
		if err != nil {
			return false, err
		}

		if !next {
			break
		}

		results = append(results, plan.Values())
	}
	p.doc[p.childFields[index]] = results
	return true, nil
}

/*

query {
	user {
		_key
		name
		points
		verified

		_version {
			cid
		}
	}
}

scan node
=========
{
	_key: bae-ALICE,
	name: Alice,
	points: 124,
	verified: false
}

_version: commitSelectTopNode(append)
===================
[
	{
		cid: QmABC
	},
	{
		cid: QmDEF
	}
	...
]


*/

func (p *parallelNode) Values() map[string]interface{} {
	return p.doc
}

func (p *parallelNode) Source() planNode { return p.multiscan }

func (p *parallelNode) Children() []planNode {
	return p.children
}

func (p *parallelNode) AddChild(field string, node planNode) error {
	p.children = append(p.children, node)
	p.childFields = append(p.childFields, field)
	return nil
}

/*
user {
	friends {
		name
	}

	addresses {
		street_name
	}
}

Select {
	source: scanNode(user)
}

		||||||
		\/\/\/

Select {
	source: TypeJoin(friends, user) {
		joinPlan {
			typeJoinMany {
				root: scanNode(user)
				subType: Select {
					source: scanNode(friends)
				}
			}
		}
	},
}

		||||||
		\/\/\/

Select {
	source: MultiNode[
		{
			TypeJoin(friends, user) {
				joinPlan {
					typeJoinMany {
						root: multiscan(scanNode(user))
						subType: Select {
							source: scanNode(friends)
						}
					}
				}
			}
		},
		{
			TypeJoin(addresses, user) {
				joinPlan {
					typeJoinMany {
						root: multiscan(scanNode(user))
						subType: Select {
							source: scanNode(addresses)
						}
					}
				}
			}
		}]
	}
}

select addSubPlan {
	check if source is MultiNode
	yes =>
		get multiScan node
		create new plan with multi scan node
		append
	no = >
		create new multinode
		get scan node from existing source
		create multiscan
		replace existing source scannode with multiScan
		add existing source to new MultiNode
		add new plan to multNode

}

Select {
	source: Parallel {[
		TypeJoin {

		},
		commitScan {

		}
	]}
}
*/

// @todo: Document AddSubPlan method
func (s *selectNode) addSubPlan(field string, plan planNode) error {
	src := s.source
	switch node := src.(type) {
	// if its a scan node, we either replace or create a multinode
	case *scanNode, *pipeNode:
		switch plan.(type) {
		case mergeNode:
			s.source = plan
		case appendNode:
			m := &parallelNode{p: s.p, doc: make(map[string]interface{})}
			if err := m.AddChild("", src); err != nil {
				return err
			}
			if err := m.AddChild(field, plan); err != nil {
				return err
			}
			s.source = m
		default:
			return errors.New("Sub plan needs to be either a MergeNode or an AppendNode")
		}

	// source is a mergeNode, like a TypeJoin
	case mergeNode:
		origScan := s.p.walkAndFindPlanType(plan, &scanNode{}).(*scanNode)
		if origScan == nil {
			return errors.New("Failed to find original scan node in plan graph")
		}
		// create our new multiscanner
		multiscan := &multiScanNode{scanNode: origScan}
		// create multinode
		multinode := &parallelNode{
			p:         s.p,
			multiscan: multiscan,
			doc:       make(map[string]interface{}),
		}
		// replace our current source internal scanNode with our new multiscanner
		if err := s.p.walkAndReplacePlan(src, origScan, multiscan); err != nil {
			return err
		}
		// add our newly updated source to the multinode
		if err := multinode.AddChild("", src); err != nil {
			return err
		}
		multiscan.addReader()
		// replace our new node internal scanNode with our new multiscanner
		if err := s.p.walkAndReplacePlan(plan, origScan, multiscan); err != nil {
			return err
		}
		// add our newly updated plan to the multinode
		if err := multinode.AddChild(field, plan); err != nil {
			return err
		}
		multiscan.addReader()
		s.source = multinode

	// we already have an existing MultiNode as our source
	case MultiNode:
		switch plan.(type) {
		// easy, just append, since append doest need any internal relaced scannode
		case appendNode:
			if err := node.AddChild(field, plan); err != nil {
				return err
			}

		// We have a internal multiscanNode on our MultiNode
		case mergeNode:
			multiscan, sourceIsMultiscan := node.Source().(*multiScanNode)
			if !sourceIsMultiscan {
				return errors.New("Merge node source must be a multiScanNode")
			}

			// replace our new node internal scanNode with our existing multiscanner
			if err := s.p.walkAndReplacePlan(plan, multiscan.Source(), multiscan); err != nil {
				return err
			}
			multiscan.addReader()
			// add our newly updated plan to the multinode
			if err := node.AddChild(field, plan); err != nil {
				return err
			}
		default:
			return errors.New("Sub plan needs to be either a MergeNode or an AppendNode")
		}
	}
	return nil
}
