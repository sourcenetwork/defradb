// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"
	"encoding/binary"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

var (
	keysNs         = "k" // /keys namespace /set/k/<key>/{v,p}
	valueSuffix    = "v" // value key
	prioritySuffix = "p" // priority key
)

// baseCRDT is embedded as a base layer into all
// the core CRDT implementations to reduce code
// duplication, and better manage the overhead
// tasks that all the CRDTs need to implement anyway
type baseCRDT struct {
	store          core.DSReaderWriter
	namespace      ds.Key
	keysNs         string
	valueSuffix    string
	prioritySuffix string
}

// @todo paramaterize ns/suffix
func newBaseCRDT(store core.DSReaderWriter, namespace ds.Key) baseCRDT {
	return baseCRDT{
		store:          store,
		namespace:      namespace,
		keysNs:         keysNs,
		valueSuffix:    valueSuffix,
		prioritySuffix: prioritySuffix,
	}
}

func (base baseCRDT) keyPrefix(key string) ds.Key {
	return base.namespace.ChildString(key)
}

func (base baseCRDT) valueKey(key string) ds.Key {
	return base.namespace.ChildString(key).Instance(base.valueSuffix)
}

func (base baseCRDT) priorityKey(key string) ds.Key {
	return base.namespace.ChildString(key).Instance(base.prioritySuffix)
}

// Commented because this function is unused (for linter).
// func (base baseCRDT) typeKey(key string) ds.Key {
// 	return base.namespace.ChildString(key).Instance(crdtTypeSuffix)
// }

// func (base baseCRDT) dataTypeKey(key string) ds.Key {
// 	return base.namespace.ChildString(key).Instance(dataTypeSuffix)
// }

func (base baseCRDT) setPriority(ctx context.Context, key string, priority uint64) error {
	prioK := base.priorityKey(key)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority+1)
	if n == 0 {
		return errors.New("error encoding priority")
	}

	return base.store.Put(ctx, prioK, buf[0:n])
}

// get the current priority for given key
func (base baseCRDT) getPriority(ctx context.Context, key string) (uint64, error) {
	pKey := base.priorityKey(key)
	pbuf, err := base.store.Get(ctx, pKey)
	if err != nil {
		if err == ds.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	prio, num := binary.Uvarint(pbuf)
	if num <= 0 {
		return 0, errors.New("failed to decode priority")
	}
	return prio, nil
}
