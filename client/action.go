// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// Action is a possible action that Defra can execute.
type Action uint16

const (
	// There is no safe default `Action`, so skip `0`

	TruncateAction         Action = 1
	RefreshDatastoreAction Action = 2
)

// ActionStatus describes the current status of an action execution.
type ActionStatus uint16

const (
	NoneActionStatus       ActionStatus = 0
	InProgressActionStatus ActionStatus = 1
	ErroredActionStatus    ActionStatus = 2
	CompletedActionStatus  ActionStatus = 2
)

// ActionExecution describes an action execution.
type ActionExecution struct {
	// The ID of the collection that this action is executing on.
	CollectionID string

	// The current action being executed.
	Action Action

	// The current status of this action execution.
	Status ActionStatus
}
