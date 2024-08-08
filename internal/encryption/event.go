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
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
)

const RequestKeysEventName = event.Name("enc-keys-request")
const KeysRetrievedEventName = event.Name("enc-keys-retrieved")

// RequestKeysEvent represents a request of a node to fetch an encryption key for a specific
// docID/field
//
// It must only contain public elements not protected by ACP.
type RequestKeysEvent struct {
	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Keys is a list of the keys that are being requested.
	Keys []core.EncStoreDocKey
}

// RequestedKeyEventData represents the data that was retrieved for a specific key.
type RequestedKeyEventData struct {
	// Key is the encryption key that was retrieved.
	Key []byte
}

// KeyRetrievedEvent represents a key that was retrieved.
type KeyRetrievedEvent struct {
	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Keys is a map of the requested keys to the corresponding raw encryption keys.
	Keys map[core.EncStoreDocKey][]byte
}

// NewRequestKeysMessage creates a new event message for a request of a node to fetch an encryption key
// for a specific docID/field
func NewRequestKeysMessage(
	schemaRoot string,
	keys []core.EncStoreDocKey,
) event.Message {
	return event.NewMessage(RequestKeysEventName, RequestKeysEvent{
		SchemaRoot: schemaRoot,
		Keys:       keys,
	})
}

// NewKeysRetrievedMessage creates a new event message for a key that was retrieved
func NewKeysRetrievedMessage(
	schemaRoot string,
	keys map[core.EncStoreDocKey][]byte,
) event.Message {
	return event.NewMessage(KeysRetrievedEventName, KeyRetrievedEvent{
		SchemaRoot: schemaRoot,
		Keys:       keys,
	})
}
