// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package secore

// ArtifactType represents the type of SE artifact
type ArtifactType string

// OperationType represents the type of operation
type OperationType string

const (
	// ArtifactTypeEqualityTag represents an equality search tag
	ArtifactTypeEqualityTag ArtifactType = "equality_tag"

	// OperationAdd represents an add operation
	OperationAdd OperationType = "add"
	// OperationDelete represents a delete operation
	OperationDelete OperationType = "delete"
)

// Artifact represents a searchable encryption operation to be replicated
type Artifact struct {
	Type         ArtifactType
	CollectionID string
	FieldName    string
	DocID        string
	Tag          []byte
	Operation    OperationType
}