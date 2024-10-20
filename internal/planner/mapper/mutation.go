// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

type MutationType int

const (
	NoneMutationType MutationType = iota
	CreateObjects
	UpdateObjects
	DeleteObjects
	UpsertObjects
)

// Mutation represents a request to mutate data stored in Defra.
type Mutation struct {
	// The underlying Select, defining the information requested upon return.
	Select

	// The type of mutation. For example a create request.
	Type MutationType

	// CreateInput is the array of maps of fields and values used for a create mutation.
	CreateInput []map[string]any

	// UpdateInput is a map of fields and values used for an update mutation.
	UpdateInput map[string]any

	// Encrypt is a flag to indicate if the input data should be encrypted.
	Encrypt bool

	// EncryptFields is a list of fields from the input data that should be encrypted.
	EncryptFields []string
}
