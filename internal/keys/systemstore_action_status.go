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
// It is stored in the following format:
// [CollectionID]/[Action] => [ActionStatus]
type ActionStatusKey struct {
	CollectionID string
	Action       client.Action
}

var _ Key = (*CollectionKey)(nil)

// Returns a formatted collection key for the system data store.
// It assumes the id of the collection is non-empty.
func NewActionStatusKey(collectionID string, action client.Action) ActionStatusKey {
	return ActionStatusKey{
		CollectionID: collectionID,
		Action:       action,
	}
}

func NewEmptyActionStatusKey() ActionStatusKey {
	return ActionStatusKey{}
}

func NewActionStatusKeyString(keyString string) (ActionStatusKey, error) {
	keyString = strings.TrimPrefix(keyString, ACTION_STATUS+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return ActionStatusKey{}, ErrInvalidKey
	}

	action, err := strconv.Atoi(elements[1])
	if err != nil {
		return ActionStatusKey{}, err
	}

	return ActionStatusKey{
		CollectionID: elements[0],
		Action:       client.Action(action),
	}, nil
}

func (k ActionStatusKey) ToString() string {
	if k.CollectionID == "" {
		return fmt.Sprintf("%s/", ACTION_STATUS)
	}

	return fmt.Sprintf("%s/%s/%v", ACTION_STATUS, k.CollectionID, k.Action)
}

func (k ActionStatusKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ActionStatusKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
