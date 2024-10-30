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
	"testing"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func newDS() datastore.DSReaderWriter {
	return datastore.AsDSReaderWriter(ds.NewMapDatastore())
}

func TestBaseCRDTvalueKey(t *testing.T) {
	vk := keys.DataStoreKey{}.WithDocID("mykey").WithValueFlag()
	if vk.ToString() != "/v/mykey" {
		t.Errorf("Incorrect valueKey. Have %v, want %v", vk.ToString(), "/v/mykey")
	}
}

func TestBaseCRDTprioryKey(t *testing.T) {
	pk := keys.DataStoreKey{}.WithDocID("mykey").WithPriorityFlag()
	if pk.ToString() != "/p/mykey" {
		t.Errorf("Incorrect priorityKey. Have %v, want %v", pk.ToString(), "/p/mykey")
	}
}

func TestBaseCRDTSetGetPriority(t *testing.T) {
	store := newDS()

	ctx := context.Background()
	err := setPriority(ctx, store, keys.DataStoreKey{}.WithDocID("mykey"), 10)
	if err != nil {
		t.Errorf("baseCRDT failed to set Priority. err: %v", err)
		return
	}

	priority, err := getPriority(ctx, store, keys.DataStoreKey{}.WithDocID("mykey"))
	if err != nil {
		t.Errorf("baseCRDT failed to get priority. err: %v", err)
		return
	}

	if priority != uint64(10) {
		t.Errorf("baseCRDT incorrect priority. Have %v, want %v", priority, uint64(10))
	}
}
