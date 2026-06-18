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

// ActionStatusKey points to the current status of an action execution.
//
// It is stored in one of the following formats:
//
//	[CollectionID]/[Action]           => [ActionStatus]   // collection-wide actions
//	[CollectionID]/[Action]/[Subject] => [ActionStatus]   // per-subject actions
//
// The optional Subject segment distinguishes concurrent executions of the same action on one
// collection (for example two index builds keyed by index ID). Collection-wide actions
// (truncate, datastore refresh) leave Subject empty.
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

// CollectionPrefix returns the byte prefix covering every action record for the key's
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
	if k.CollectionID == "" {
		return fmt.Sprintf("%s/", ACTION_STATUS)
	}

	if k.Subject != "" {
		return fmt.Sprintf("%s/%s/%v/%s", ACTION_STATUS, k.CollectionID, k.Action, k.Subject)
	}

	return fmt.Sprintf("%s/%s/%v", ACTION_STATUS, k.CollectionID, k.Action)
}

func (k ActionStatusKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ActionStatusKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
