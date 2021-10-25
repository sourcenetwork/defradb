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
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/core"
)

func newDS() core.DSReaderWriter {
	return ds.NewMapDatastore()
}

func newSeededDS() core.DSReaderWriter {
	return newDS()
}

func exampleBaseCRDT() baseCRDT {
	return newBaseCRDT(newSeededDS(), core.DataStoreKey{})
}

func TestBaseCRDTNew(t *testing.T) {
	base := newBaseCRDT(newDS(), core.DataStoreKey{})
	if base.store == nil {
		t.Error("newBaseCRDT needs to init store")
	}
}

func TestBaseCRDTvalueKey(t *testing.T) {
	base := exampleBaseCRDT()
	vk := base.key.WithDocKey("mykey").WithValueFlag()
	if vk.ToString() != "/mykey:v" {
		t.Errorf("Incorrect valueKey. Have %v, want %v", vk.ToString(), "/mykey:v")
	}
}

func TestBaseCRDTprioryKey(t *testing.T) {
	base := exampleBaseCRDT()
	pk := base.key.WithDocKey("mykey").WithPriorityFlag()
	if pk.ToString() != "/mykey:p" {
		t.Errorf("Incorrect priorityKey. Have %v, want %v", pk.ToString(), "/mykey:p")
	}
}

func TestBaseCRDTSetGetPriority(t *testing.T) {
	base := exampleBaseCRDT()
	err := base.setPriority(base.key.WithDocKey("mykey"), 10)
	if err != nil {
		t.Errorf("baseCRDT failed to set Priority. err: %v", err)
		return
	}

	priority, err := base.getPriority(base.key.WithDocKey("mykey"))
	if err != nil {
		t.Errorf("baseCRDT failed to get priority. err: %v", err)
		return
	}

	if priority-1 != uint64(10) {
		t.Errorf("baseCRDT incorrect priority. Have %v, want %v", priority, uint64(10))
	}
}
