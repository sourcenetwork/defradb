// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package event

import (
	"github.com/ipfs/go-cid"
)

// Name identifies an event
type Name string

const (
	// WildCardName is the alias used to subscribe to all events.
	WildCardName = Name("*")
	// MergeName is the name of the net merge request event.
	MergeName = Name("merge")
	// MergeCompleteName is the name of the database merge complete event.
	MergeCompleteName = Name("merge-complete")
	// UpdateName is the name of the database update event.
	UpdateName = Name("update")
	// PurgeName is the name of the purge event.
	PurgeName = Name("purge")
)

// Update represents a new DAG node added to the append-only composite MerkleCRDT Clock graph
// of a document.
//
// It must only contain public elements not protected by ACP.
type Update struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Block is the encoded contents of this composite commit, it contains the Cids of the field level commits that
	// also formed this update.
	Block []byte

	// IsRetry is true if this update is a retry of a previously failed update.
	IsRetry bool

	// Success is a channel that will receive a boolean value indicating if the update was successful.
	// It is used during retries.
	Success chan bool
}

// Merge is a notification that a merge can be performed up to the provided CID.
type Merge struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string
}

// MergeComplete is a notification that a merge has been completed.
type MergeComplete struct {
	// Merge is the merge that was completed.
	Merge Merge

	// Decrypted specifies if the merge payload was decrypted.
	Decrypted bool
}

// Message contains event info.
type Message struct {
	// Name is the name of the event this message was generated from.
	Name Name

	// Data contains optional event information.
	Data any
}

// NewMessage returns a new message with the given name and optional data.
func NewMessage(name Name, data any) Message {
	return Message{name, data}
}

// Subscription is a read-only event stream.
type Subscription struct {
	id     uint64
	value  chan Message
	events []Name
}

// Message returns the next event value from the subscription.
func (s *Subscription) Message() <-chan Message {
	return s.value
}
