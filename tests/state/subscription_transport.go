// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package state

// SubscriptionTransportType specifies the transport protocol to use for GraphQL subscriptions.
type SubscriptionTransportType string

const (
	// SSETransportType uses Server-Sent Events for subscriptions (default).
	SSETransportType SubscriptionTransportType = "sse"
	// WebSocketTransportType uses WebSocket for subscriptions.
	WebSocketTransportType SubscriptionTransportType = "websocket"
)
