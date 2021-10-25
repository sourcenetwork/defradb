// Copyright 2020 Source Inc.
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
	"encoding/binary"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

// baseCRDT is embedded as a base layer into all
// the core CRDT implementations to reduce code
// duplcation, and better manage the overhead
// tasks that all the CRDTs need to implement anyway
type baseCRDT struct {
	store core.DSReaderWriter
	key   core.DataStoreKey
}

// @TODO paramaterize ns/suffix
func newBaseCRDT(store core.DSReaderWriter, key core.DataStoreKey) baseCRDT {
	return baseCRDT{
		store: store,
		key:   key,
	}
}

func (base baseCRDT) setPriority(key core.DataStoreKey, priority uint64) error {
	prioK := key.WithPriorityFlag()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority+1)
	if n == 0 {
		return errors.New("error encoding priority")
	}

	return base.store.Put(prioK.ToDS(), buf[0:n])
}

// get the current priority for given key
func (base baseCRDT) getPriority(key core.DataStoreKey) (uint64, error) {
	pKey := key.WithPriorityFlag()
	pbuf, err := base.store.Get(pKey.ToDS())
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
