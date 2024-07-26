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

var (
	_ planNode = (*averageNode)(nil)
	_ planNode = (*countNode)(nil)
	_ planNode = (*createNode)(nil)
	_ planNode = (*dagScanNode)(nil)
	_ planNode = (*deleteNode)(nil)
	_ planNode = (*groupNode)(nil)
	_ planNode = (*limitNode)(nil)
	_ planNode = (*multiScanNode)(nil)
	_ planNode = (*orderNode)(nil)
	_ planNode = (*parallelNode)(nil)
	_ planNode = (*pipeNode)(nil)
	_ planNode = (*scanNode)(nil)
	_ planNode = (*selectNode)(nil)
	_ planNode = (*selectTopNode)(nil)
	_ planNode = (*sumNode)(nil)
	_ planNode = (*topLevelNode)(nil)
	_ planNode = (*typeIndexJoin)(nil)
	_ planNode = (*typeJoinMany)(nil)
	_ planNode = (*typeJoinOne)(nil)
	_ planNode = (*updateNode)(nil)
	_ planNode = (*valuesNode)(nil)
	_ planNode = (*viewNode)(nil)
	_ planNode = (*lensNode)(nil)
	_ planNode = (*operationNode)(nil)

	_ MultiNode = (*parallelNode)(nil)
	_ MultiNode = (*topLevelNode)(nil)
	_ MultiNode = (*operationNode)(nil)
)

// type joinNode struct {
// 	p *Planner
// }
//
// // scatter group by or aggregate operations
// type scatterNode struct {
// 	p *Planner
// }
//
// // apply an aggregate function to a result
// type aggregateNode struct {
// 	p *Planner
// }
//
// // apply a "Having" operation
// type filterHavingNode struct {
// 	p *Planner
// }
//
// // noop
// type noopNode struct {
// 	p *Planner
// }
//
// // parallel planner, that is used to execute multiple plan trees in parallel.
// type parallelNode struct {
// 	pNodes []planNode
// }
