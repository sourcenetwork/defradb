// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

type EncryptedIndexType string

const (
	EncryptedIndexTypeEquality EncryptedIndexType = "equality"
)

// EncryptedIndexDescription represents metadata for an encrypted index.
type EncryptedIndexDescription struct {
	// FieldName is the name of the field that is being indexed.
	FieldName string

	// Type is the type of searchable encryption.
	// Currently only "equality" is supported.
	Type EncryptedIndexType
}

// EncryptedIndexCreateRequest describes an encrypted index creation request.
type EncryptedIndexCreateRequest struct {
	// FieldName contains the name of the field that is being indexed.
	FieldName string
	// Type is the type of searchable encryption.
	// Currently only "equality" is supported.
	Type EncryptedIndexType
}
