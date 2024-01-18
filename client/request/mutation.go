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

import "github.com/sourcenetwork/immutable"

type MutationType int

const (
	NoneMutationType = MutationType(iota)
	CreateObjects
	UpdateObjects
	DeleteObjects
)

// ObjectMutation is a field on the `mutation` operation of a graphql request. It includes
// all the possible arguments.
type ObjectMutation struct {
	Field
	Type MutationType

	// Collection is the target collection name
	// if this mutation is on an object.
	Collection string

	IDs    immutable.Option[[]string]
	Filter immutable.Option[Filter]
	Input  map[string]any

	Fields []Selection
}

// ToSelect returns a basic Select object, with the same Name, Alias, and Fields as
// the Mutation object. Used to create a Select planNode for the mutation return objects.
func (m ObjectMutation) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  m.Collection,
			Alias: m.Alias,
		},
		Fields: m.Fields,
		DocIDs: m.IDs,
		Filter: m.Filter,
	}
}
