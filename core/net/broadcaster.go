// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import "time"

// A Broadcaster provides a way to send (notify) an opaque payload to
// all replicas and to retrieve payloads broadcasted.
type Broadcaster interface {
	// Send broadcasts a message without blocking
	Send(v any) error

	// SendWithTimeout broadcasts a message, blocks up to timeout duration
	SendWithTimeout(v any, d time.Duration) error
}
