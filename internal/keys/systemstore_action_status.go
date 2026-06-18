// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"fmt"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
)

// An action execution is stored across three sibling key spaces, kept separate so a reader that
// only needs the status does not load the reason or payload:
//
//	/a/s => ActionStatus
//	/a/r => Reason (errored actions)
//	/a/p => Payload (action-specific opaque bytes)
//
// The optional Subject segment distinguishes concurrent executions of the same action on one
// collection; collection-wide actions (truncate, datastore refresh) leave it empty.

// actionKeyString formats an action key under the given prefix.
func actionKeyString(prefix, collectionID string, action client.Action, subject string) string {
	if collectionID == "" {
		return fmt.Sprintf("%s/", prefix)
	}
	if subject != "" {
		return fmt.Sprintf("%s/%s/%v/%s", prefix, collectionID, action, subject)
	}
	return fmt.Sprintf("%s/%s/%v", prefix, collectionID, action)
}

// ActionStatusKey points to the current status of an action execution.
type ActionStatusKey struct {
	CollectionID string
	Action       client.Action
	Subject      string
}

var _ Key = (*ActionStatusKey)(nil)

// Returns a formatted collection key for the system data store.
// It assumes the id of the collection is non-empty.
func NewActionStatusKey(collectionID string, action client.Action) ActionStatusKey {
	return ActionStatusKey{
		CollectionID: collectionID,
		Action:       action,
	}
}

// NewActionStatusSubjectKey returns a key for a per-subject action execution.
func NewActionStatusSubjectKey(collectionID string, action client.Action, subject string) ActionStatusKey {
	return ActionStatusKey{
		CollectionID: collectionID,
		Action:       action,
		Subject:      subject,
	}
}

func NewEmptyActionStatusKey() ActionStatusKey {
	return ActionStatusKey{}
}

// CollectionPrefix returns the byte prefix covering every action status record for the key's
// collection, across all actions and subjects.
func (k ActionStatusKey) CollectionPrefix() []byte {
	return []byte(ACTION_STATUS + "/" + k.CollectionID + "/")
}

func NewActionStatusKeyString(keyString string) (ActionStatusKey, error) {
	keyString = strings.TrimPrefix(keyString, ACTION_STATUS+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 && len(elements) != 3 {
		return ActionStatusKey{}, ErrInvalidKey
	}

	action, err := strconv.Atoi(elements[1])
	if err != nil {
		return ActionStatusKey{}, err
	}

	key := ActionStatusKey{
		CollectionID: elements[0],
		Action:       client.Action(action),
	}
	if len(elements) == 3 {
		key.Subject = elements[2]
	}
	return key, nil
}

func (k ActionStatusKey) ToString() string {
	return actionKeyString(ACTION_STATUS, k.CollectionID, k.Action, k.Subject)
}

func (k ActionStatusKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ActionStatusKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// ActionReasonKey points to the reason an action errored.
type ActionReasonKey struct {
	CollectionID string
	Action       client.Action
	Subject      string
}

var _ Key = (*ActionReasonKey)(nil)

// NewActionReasonKey returns a reason key for the given action execution.
func NewActionReasonKey(collectionID string, action client.Action, subject string) ActionReasonKey {
	return ActionReasonKey{
		CollectionID: collectionID,
		Action:       action,
		Subject:      subject,
	}
}

func (k ActionReasonKey) ToString() string {
	return actionKeyString(ACTION_REASON, k.CollectionID, k.Action, k.Subject)
}

func (k ActionReasonKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ActionReasonKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// ActionPayloadKey points to an action's opaque, action-specific payload (for example an index
// build's watermark). Each action defines its own encoding.
type ActionPayloadKey struct {
	CollectionID string
	Action       client.Action
	Subject      string
}

var _ Key = (*ActionPayloadKey)(nil)

// NewActionPayloadKey returns a payload key for the given action execution.
func NewActionPayloadKey(collectionID string, action client.Action, subject string) ActionPayloadKey {
	return ActionPayloadKey{
		CollectionID: collectionID,
		Action:       action,
		Subject:      subject,
	}
}

func (k ActionPayloadKey) ToString() string {
	return actionKeyString(ACTION_PAYLOAD, k.CollectionID, k.Action, k.Subject)
}

func (k ActionPayloadKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ActionPayloadKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
