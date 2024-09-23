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
)

// ObjectMutation is a field on the `mutation` operation of a graphql request. It includes
// all the possible arguments.
type ObjectMutation struct {
	Field
	ChildSelect

	Filterable
	DocIDsFilter

	// Type is the type of mutatation that this object represents.
	//
	// For example [CreateObjects].
	Type MutationType

	// Collection is the target collection name.
	Collection string

	// Input is the array of json representations of the fieldName-value pairs of document
	// properties to mutate.
	//
	// This is ignored for [DeleteObjects] mutations.
	Input []map[string]any

	// Encrypt is a boolean flag that indicates whether the input data should be encrypted.
	Encrypt bool

	// EncryptFields is a list of doc fields from input data that should be encrypted.
	EncryptFields []string
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
