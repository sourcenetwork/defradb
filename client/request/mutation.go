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

type MutationType int

const (
	NoneMutationType = MutationType(iota)
	CreateObjects
	UpdateObjects
	DeleteObjects
	UpsertObjects
)

// ObjectMutation is a field on the `mutation` operation of a graphql request. It includes
// all the possible arguments.
type ObjectMutation struct {
	Filterable

	// UpdateInput is a map of fields and values used for an update mutation.
	UpdateInput map[string]any

	Field

	// Collection is the target collection name.
	Collection string

	DocIDsFilter

	ChildSelect

	// CreateInput is the array of maps of fields and values used for a create mutation.
	CreateInput []map[string]any

	// EncryptFields is a list of doc fields from input data that should be encrypted.
	EncryptFields []string

	// Type is the type of mutatation that this object represents.
	//
	// For example [CreateObjects].
	Type MutationType

	// Encrypt is a boolean flag that indicates whether the input data should be encrypted.
	Encrypt bool
}

// ToSelect returns a basic Select object, with the same Name, Alias, and Fields as
// the Mutation object. Used to create a Select planNode for the mutation return objects.
func (m ObjectMutation) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  m.Collection,
			Alias: m.Alias,
		},
		ChildSelect:  m.ChildSelect,
		DocIDsFilter: m.DocIDsFilter,
		Filterable:   m.Filterable,
	}
}
