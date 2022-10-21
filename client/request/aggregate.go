// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import "github.com/sourcenetwork/defradb/client"

type Aggregate struct {
	Field

	Targets []*AggregateTarget
}

type AggregateTarget struct {
	HostName  string
	ChildName client.Option[string]

	Limit   client.Option[uint64]
	Offset  client.Option[uint64]
	OrderBy client.Option[OrderBy]
	Filter  client.Option[Filter]
}
