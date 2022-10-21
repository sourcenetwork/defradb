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

type MutationType int

const (
	NoneMutationType = MutationType(iota)
	CreateObjects
	UpdateObjects
	DeleteObjects
)

// Mutation is a field on the MutationType
// of a graphql query. It includes all the possible
// arguments and all
//
// @todo: Change name to ObjectMutation to indicate
// generated object mutation actions
type Mutation struct {
	Field
	Type MutationType

	// Schema is the target schema/collection
	// if this mutation is on an object.
	Schema string

	IDs    client.Option[[]string]
	Filter client.Option[Filter]
	Data   string

	Fields []Selection
}

// ToSelect returns a basic Select object, with the same Name, Alias, and Fields as
// the Mutation object. Used to create a Select planNode for the mutation return objects.
func (m Mutation) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  m.Schema,
			Alias: m.Alias,
		},
		Fields:  m.Fields,
		DocKeys: m.IDs,
		Filter:  m.Filter,
	}
}
