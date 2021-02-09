package planner

import (
	"errors"

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
	AddChild(planNode) error
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
// However, they are entirely independant graphs, so they can
// be executed in parallel.
//
type parallelNode struct { // serialNode?
	p *Planner

	children []planNode
	childID  []string

	multiscan *multiScanNode

	doc map[string]interface{}
}

// func (n *selectTopNode) Init() error                    { return n.plan.Init() }
// func (n *selectTopNode) Start() error                   { return n.plan.Start() }
// func (n *selectTopNode) Next() (bool, error)            { return n.plan.Next() }
// func (n *selectTopNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
// func (n *selectTopNode) Values() map[string]interface{} { return n.plan.Values() }
// func (n *selectTopNode) Close() {
// 	if n.plan != nil {
// 		n.plan.Close()
// 	}
// }

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
	p.applyToPlans(func(n planNode) error {
		n.Spans(spans)
		return nil
	})
}

func (p *parallelNode) Close() {
	p.applyToPlans(func(n planNode) error {
		n.Close()
		return nil
	})
}

// Next loops through all the children nodes, and calls Next().
// It only needs a single child plan to return true for it
// to return true. Same with errors.
func (p *parallelNode) Next() (bool, error) {
	var orNext bool
	for _, plan := range p.children {
		next, err := plan.Next()
		if err != nil {
			return false, err
		}

		// logical OR all the next results together
		orNext = orNext || next
	}
	// if none of the children return true for next, then this will be false.
	// if ANY of the children return true, this will be true (logical OR)
	return orNext, nil
}

func (p *parallelNode) Values() map[string]interface{} {
	var result map[string]interface{}
	for _, plan := range p.children {
		if doc := plan.Values(); doc != nil {
			result = doc
		}
	}

	return result
}

func (p *parallelNode) Source() planNode { return p.multiscan }

func (p *parallelNode) Children() []planNode {
	return p.children
}

func (p *parallelNode) AddChild(node planNode) error {
	p.children = append(p.children, node)
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

func (s *selectNode) addSubPlan(plan planNode) error {
	var multinode MultiNode
	var multiscan *multiScanNode
	switch src := s.source.(type) {
	case MultiNode:
		// multiscan, ok := src.Source().(*multiScanNode)
		// if !ok {
		// 	return nil, errors.New("Exisint MultiNode doesn't have a multiscan")
		// }
		// origScan := multiscan.scanNode
		// p.walkAndReplaceNode(plan, origScan, multiscan)
		// return src.AddChild(field, plan)
		multinode = src
		multiscan = multinode.Source().(*multiScanNode)
	case *scanNode:
		// we have a simple scanNode as our source
		// no need to do anything with the MultiNodes
		// just set the source to the target plan, and exit
		s.source = plan
		return nil
	default: // no existing multinode, and our current source is a complex non scanNode node.
		// get original scanNode
		origScan := s.p.walkAndFindPlanType(plan, &scanNode{}).(*scanNode)
		if origScan == nil {
			return errors.New("Failed to find original scan node in plan graph")
		}
		// create our new multiscanner
		multiscan = &multiScanNode{scanNode: origScan}
		multiscan.addReader()
		// create multinode
		multinode = &parallelNode{
			multiscan: multiscan,
		}
		// replace our current source internal scanNode with our new multiscanner
		if err := s.p.walkAndReplacePlan(src, origScan, multiscan); err != nil {
			return err
		}
		// add our newly updated source to the multinode
		if err := multinode.AddChild(src); err != nil {
			return err
		}
		s.source = multinode
	}

	// if we've got here, then we have an instanciated multinode ready to add
	// our new plan to
	multiscan.addReader()
	scan := multiscan.Source()
	// replace our current source internal scanNode with our new multiscanner
	if err := s.p.walkAndReplacePlan(plan, scan, multiscan); err != nil {
		return err
	}
	// add our newly updated source to the multinode
	return multinode.AddChild(plan)
}

// func (p *Planner) parallelNode() (*parallelNode, error) {
// 	mp := &parallelNode{
// 		children: nodes,
// 	}
// }
