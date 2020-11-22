package planner

type joinNode struct {
	p *Planner
}

// applys a 'Group By' operation
type groupNode struct {
	p *Planner
} // gatherNode?

// scatter group by or aggregate operations
type scatterNode struct {
	p *Planner
}

// apply an aggregate function to a result
type aggregateNode struct {
	p *Planner
}

// // apply a "Having" operation
// type filterHavingNode struct {
// 	p *Planner
// }

// order the results
type sortNode struct {
	p *Planner
}

// limit the results
type limitNode struct {
	p *Planner
}

// the final field select and render
type renderNode struct { // selectNode??
	p *Planner
}

// noop
type noopNode struct {
	p *Planner
}

// parellel planner, that is used to execute multiple plan trees in parallel.
type parallelNode struct {
	pNodes []planNode
}
