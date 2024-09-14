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
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/event"
)

const RequestKeysEventName = event.Name("enc-keys-request")

// RequestKeysEvent represents a request of a node to fetch an encryption key for a specific
// docID/field
//
// It must only contain public elements not protected by ACP.
type RequestKeysEvent struct {
	// Keys is a list of the keys that are being requested.
	Keys []cidlink.Link

	Resp chan<- Result
}

// RequestedKeyEventData represents the data that was retrieved for a specific key.
type RequestedKeyEventData struct {
	// Key is the encryption key that was retrieved.
	Key []byte
}

// KeyRetrievedEvent represents a key that was retrieved.
type Item struct {
	Link  []byte
	Block []byte
}

type Result struct {
	Items []Item
	Error error
}

type Results struct {
	output chan Result
}

func (r *Results) Get() <-chan Result {
	return r.output
}

func NewResults() (*Results, chan<- Result) {
	ch := make(chan Result, 1)
	return &Results{
		output: ch,
	}, ch
}

// NewRequestKeysMessage creates a new event message for a request of a node to fetch an encryption key
// for a specific docID/field
func NewRequestKeysMessage(keys []cidlink.Link) (event.Message, *Results) {
	res, ch := NewResults()
	return event.NewMessage(RequestKeysEventName, RequestKeysEvent{
		Keys: keys,
		Resp: ch,
	}), res
}
