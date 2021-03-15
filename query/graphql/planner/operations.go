package planner

var (
	_ planNode = (*scanNode)(nil)
	_ planNode = (*headsetScanNode)(nil)
	_ planNode = (*limitNode)(nil)
	_ planNode = (*selectNode)(nil)
	_ planNode = (*selectTopNode)(nil)
	_ planNode = (*sortNode)(nil)
	_ planNode = (*renderNode)(nil)
	_ planNode = (*typeIndexJoin)(nil)
	_ planNode = (*typeJoinOne)(nil)
	_ planNode = (*typeJoinMany)(nil)
)

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

// noop
type noopNode struct {
	p *Planner
}

// // parellel planner, that is used to execute multiple plan trees in parallel.
// type parallelNode struct {
// 	pNodes []planNode
// }
