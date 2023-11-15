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
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

func newStores() datastore.MultiStore {
	root := ds.NewMapDatastore()
	return datastore.MultiStoreFrom(root)
}

func TestNewBlankFactory(t *testing.T) {
	f := NewFactory(nil)
	if f == nil {
		t.Fatal("Returned factory is a nil pointer")
	}
}

func TestNewFactoryWithStores(t *testing.T) {
	m := newStores()
	f := NewFactory(m)
	if f == nil {
		t.Fatal("Returned factory is a nil pointer")
	}

	assert.Equal(t, m.Datastore(), f.Datastore())
	assert.Equal(t, m.Headstore(), f.Headstore())
	assert.Equal(t, m.DAGstore(), f.DAGstore())
	assert.Equal(t, m.Systemstore(), f.Systemstore())
}

func TestFactoryMultiStoreInterface(t *testing.T) {
	m := newStores()
	f := NewFactory(m)
	if f == nil {
		t.Fatal("Returned factory is a nil pointer")
	}

	// check interface implement
	var _ datastore.MultiStore = f
	// ms = f

	// check interface functions
	assert.Equal(t, m.Datastore(), f.Datastore())
	assert.Equal(t, m.Headstore(), f.Headstore())
	assert.Equal(t, m.DAGstore(), f.DAGstore())
	assert.Equal(t, m.Systemstore(), f.Systemstore())
}

func TestFactorySetStores(t *testing.T) {
	f := NewFactory(nil)
	m := newStores()
	err := f.SetStores(m)
	assert.Nil(t, err)

	assert.Equal(t, m.Datastore(), f.Datastore())
	assert.Equal(t, m.Headstore(), f.Headstore())
	assert.Equal(t, m.DAGstore(), f.DAGstore())
	assert.Equal(t, m.Systemstore(), f.Systemstore())
}

func TestFactoryWithStores(t *testing.T) {
	f := NewFactory(nil)
	m := newStores()
	f2 := f.WithStores(m)
	// assert.NotEmpty

	assert.Nil(t, f.Datastore())
	assert.Nil(t, f.Headstore())
	assert.Nil(t, f.DAGstore())

	assert.Equal(t, m.Datastore(), f2.Datastore())
	assert.Equal(t, m.Headstore(), f2.Headstore())
	assert.Equal(t, m.DAGstore(), f2.DAGstore())
	assert.Equal(t, m.Systemstore(), f2.Systemstore())
}

func TestFullFactoryRegister(t *testing.T) {
	m := newStores()
	f := NewFactory(m)
	err := f.Register(client.LWW_REGISTER, &lwwFactoryFn)
	assert.Nil(t, err)
	assert.Equal(t, &lwwFactoryFn, f.crdts[client.LWW_REGISTER])
}

func TestBlankFactoryRegister(t *testing.T) {
	f := NewFactory(nil)
	err := f.Register(client.LWW_REGISTER, &lwwFactoryFn)
	assert.Nil(t, err)
	assert.Equal(t, &lwwFactoryFn, f.crdts[client.LWW_REGISTER])
}

func TestWithStoresFactoryRegister(t *testing.T) {
	f := NewFactory(nil)
	f.Register(client.LWW_REGISTER, &lwwFactoryFn)
	m := newStores()
	f2 := f.WithStores(m)

	assert.Equal(t, &lwwFactoryFn, f2.crdts[client.LWW_REGISTER])
}

func TestDefaultFactory(t *testing.T) {
	assert.NotNil(t, DefaultFactory)
	assert.Equal(t, &lwwFactoryFn, DefaultFactory.crdts[client.LWW_REGISTER])
}

func TestFactoryInstanceMissing(t *testing.T) {
	m := newStores()
	f := NewFactory(m)

	_, err := f.Instance(
		core.CollectionSchemaVersionKey{},
		events.EmptyUpdateChannel,
		client.LWW_REGISTER,
		core.MustNewDataStoreKey("/1/0/MyKey"),
		"",
	)
	assert.Equal(t, err, ErrFactoryTypeNoExist)
}

func TestBlankFactoryInstanceWithLWWRegister(t *testing.T) {
	m := newStores()
	f1 := NewFactory(nil)
	f1.Register(client.LWW_REGISTER, &lwwFactoryFn)
	f := f1.WithStores(m)

	crdt, err := f.Instance(
		core.CollectionSchemaVersionKey{},
		events.EmptyUpdateChannel,
		client.LWW_REGISTER,
		core.MustNewDataStoreKey("/1/0/MyKey"),
		"",
	)
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)
}

func TestBlankFactoryInstanceWithCompositeRegister(t *testing.T) {
	m := newStores()
	f1 := NewFactory(nil)
	f1.Register(client.COMPOSITE, &compFactoryFn)
	f := f1.WithStores(m)

	crdt, err := f.Instance(
		core.CollectionSchemaVersionKey{},
		events.EmptyUpdateChannel,
		client.COMPOSITE,
		core.MustNewDataStoreKey("/1/0/MyKey"),
		"",
	)
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleCompositeDAG)
	assert.True(t, ok)
}

func TestFullFactoryInstanceLWWRegister(t *testing.T) {
	m := newStores()
	f := NewFactory(m)
	f.Register(client.LWW_REGISTER, &lwwFactoryFn)

	crdt, err := f.Instance(
		core.CollectionSchemaVersionKey{},
		events.EmptyUpdateChannel,
		client.LWW_REGISTER,
		core.MustNewDataStoreKey("/1/0/MyKey"),
		"",
	)
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)
}

func TestFullFactoryInstanceCompositeRegister(t *testing.T) {
	m := newStores()
	f := NewFactory(m)
	f.Register(client.COMPOSITE, &compFactoryFn)

	crdt, err := f.Instance(
		core.CollectionSchemaVersionKey{},
		events.EmptyUpdateChannel,
		client.COMPOSITE,
		core.MustNewDataStoreKey("/1/0/MyKey"),
		"",
	)
	assert.NoError(t, err)

	_, ok := crdt.(*MerkleCompositeDAG)
	assert.True(t, ok)
}

func TestLWWRegisterFactoryFn(t *testing.T) {
	ctx := context.Background()
	m := newStores()
	f := NewFactory(m) // here factory is only needed to satisfy datastore.MultiStore interface
	crdt := lwwFactoryFn(f, core.CollectionSchemaVersionKey{}, events.EmptyUpdateChannel, "")(core.MustNewDataStoreKey("/1/0/MyKey"))

	lwwreg, ok := crdt.(*MerkleLWWRegister)
	assert.True(t, ok)

	_, _, err := lwwreg.Set(ctx, []byte("hi"))
	assert.NoError(t, err)
}

func TestCompositeRegisterFactoryFn(t *testing.T) {
	ctx := context.Background()
	m := newStores()
	f := NewFactory(m) // here factory is only needed to satisfy datastore.MultiStore interface
	crdt := compFactoryFn(f, core.CollectionSchemaVersionKey{}, events.EmptyUpdateChannel, "")(core.MustNewDataStoreKey("/1/0/MyKey"))

	merkleReg, ok := crdt.(*MerkleCompositeDAG)
	assert.True(t, ok)

	_, _, err := merkleReg.Set(ctx, []byte("hi"), []core.DAGLink{})
	assert.NoError(t, err)
}
