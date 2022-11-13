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

import "github.com/sourcenetwork/defradb/immutables"

var (
	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Field

	DocKey    immutables.Option[string]
	FieldName immutables.Option[string]
	Cid       immutables.Option[string]
	Depth     immutables.Option[uint64]

	Limit   immutables.Option[uint64]
	Offset  immutables.Option[uint64]
	OrderBy immutables.Option[OrderBy]
	GroupBy immutables.Option[GroupBy]

	Fields []Selection
}

func (c CommitSelect) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  c.Name,
			Alias: c.Alias,
		},
		Limit:   c.Limit,
		Offset:  c.Offset,
		OrderBy: c.OrderBy,
		GroupBy: c.GroupBy,
		Fields:  c.Fields,
		Root:    CommitSelection,
	}
}
