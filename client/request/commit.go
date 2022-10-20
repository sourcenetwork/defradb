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

var (
	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Field

	DocKey    client.Option[string]
	FieldName client.Option[string]
	Cid       client.Option[string]
	Depth     client.Option[uint64]

	Limit   client.Option[uint64]
	Offset  client.Option[uint64]
	OrderBy client.Option[OrderBy]
	GroupBy client.Option[GroupBy]

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
