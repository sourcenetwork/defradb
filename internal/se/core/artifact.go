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

// Artifact represents a searchable encryption artifact to be replicated to remote nodes.
// It contains the cryptographic search tag and metadata needed to store and query
// encrypted indexes on untrusted replicator nodes.
type Artifact struct {
	// CollectionID is the unique identifier of the collection
	CollectionID string

	// DocID is the unique document identifier
	DocID string

	// IndexID is the unique identifier of the encrypted index
	// Used as a domain separator in search tag computation
	IndexID string

	// SearchTag is the deterministic cryptographic tag used for searching.
	// The remote node stores this tag in its KV store at:
	// /se/<CollectionID>/<IndexID>/<SearchTag>/<DocID>
	// When querying, the client computes the same tag for the search value
	// and the remote node returns all DocIDs with matching tags.
	SearchTag []byte
}
