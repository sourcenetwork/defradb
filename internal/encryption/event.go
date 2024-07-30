// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/immutable"
)

const RequestKeyEventName = event.Name("enc-key-request")
const KeyRetrievedEventName = event.Name("enc-key-retrieved")

// RequestKeyEvent represents a request of a node to fetch an encryption key for a specific
// docID/field
//
// It must only contain public elements not protected by ACP.
type RequestKeyEvent struct {
	// DocID is the unique immutable identifier of the document that the key is being requested for.
	DocID string

	// FieldName is the name of the field for which the key is being requested.
	// If not given, the key is for the whole document.
	FieldName immutable.Option[string]

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string
}

// KeyRetrievedEvent represents a key that was retrieved.
type KeyRetrievedEvent struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// FieldName is the name of the field for which the key was retrieved.
	// If not given, the key is for the whole document.
	FieldName immutable.Option[string]

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// TODO: should be encrypted
	// Key is the encryption key that was retrieved.
	Key []byte
}

// NewRequestKeyMessage creates a new event message for a request of a node to fetch an encryption key
// for a specific docID/field
func NewRequestKeyMessage(
	docID string,
	cid cid.Cid,
	fieldName immutable.Option[string],
	schemaRoot string,
) event.Message {
	return event.NewMessage(RequestKeyEventName, RequestKeyEvent{
		DocID:      docID,
		FieldName:  fieldName,
		Cid:        cid,
		SchemaRoot: schemaRoot,
	})
}

// NewKeyRetrievedMessage creates a new event message for a key that was retrieved
func NewKeyRetrievedMessage(
	docID string,
	fieldName immutable.Option[string],
	cid cid.Cid,
	schemaRoot string,
	key []byte,
) event.Message {
	return event.NewMessage(KeyRetrievedEventName, KeyRetrievedEvent{
		DocID:      docID,
		FieldName:  fieldName,
		Cid:        cid,
		SchemaRoot: schemaRoot,
		Key:        key,
	})
}
