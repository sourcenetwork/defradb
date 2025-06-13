// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import "context"

type contextKey string

// messageChansContextKey is the context key for the messageChans
var messageChansContextKey = contextKey("messageChans")

// MessageChans contains message channels that can be used to relay important information
// after calling `cli.Start`.
type MessageChans struct {
	APIURL chan string
}

// TryGetContextMessageChans returns the message channels for the current command context
// and a boolean indicating if the message channels struct was set.
func TryGetContextMessageChans(ctx context.Context) (*MessageChans, bool) {
	node, ok := ctx.Value(messageChansContextKey).(*MessageChans)
	return node, ok
}

// SetContextMessageChans sets the message channels for the current command context.
func SetContextMessageChans(ctx context.Context, w *MessageChans) context.Context {
	return context.WithValue(ctx, messageChansContextKey, w)
}
