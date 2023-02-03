// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import immutables "github.com/sourcenetwork/immutable"

type Aggregate struct {
	Field

	Targets []*AggregateTarget
}

type AggregateTarget struct {
	HostName  string
	ChildName immutables.Option[string]

	Limit   immutables.Option[uint64]
	Offset  immutables.Option[uint64]
	OrderBy immutables.Option[OrderBy]
	Filter  immutables.Option[Filter]
}
