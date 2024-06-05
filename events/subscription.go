// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

// Subscription is a read-only event stream.
type Subscription struct {
	id     int
	value  chan any
	events []string
}

// Value returns the next event value from the subscription.
func (s *Subscription) Value() <-chan any {
	return s.value
}

// Events returns the names of all subscribed events.
func (s *Subscription) Events() []string {
	return s.events
}
