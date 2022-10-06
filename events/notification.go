// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

// Notification represents an actionable event in the subscription.
// This can either be a closing of the subscription, or an subscription item.
type Notification[T any] struct {
	hasValue bool
	value    T
}

func NewNotification[T any](value T) Notification[T] {
	return Notification[T]{
		hasValue: true,
		value:    value,
	}
}

// Closed returns true if this is a notification that the source subscription
// has been closed, otherwise true.
func (n Notification[T]) Closed() bool {
	return !n.hasValue
}

// Value returns the value of this notification.  Will be default if this is a
// close notification.
func (n Notification[T]) Value() T {
	return n.value
}
