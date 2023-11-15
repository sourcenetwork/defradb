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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

// MerkleCRDTInitFn instantiates a MerkleCRDT with a given key.
type MerkleCRDTInitFn func(core.DataStoreKey) MerkleCRDT

// MerkleCRDTFactory instantiates a MerkleCRDTInitFn with a MultiStore.
// Returns a MerkleCRDTInitFn with all the necessary stores set.
type MerkleCRDTFactory func(
	mstore datastore.MultiStore,
	schemaVersionKey core.CollectionSchemaVersionKey,
	uCh events.UpdateChannel,
	fieldName string,
) MerkleCRDTInitFn

// Factory is a helper utility for instantiating new MerkleCRDTs.
// It removes some of the overhead of having to coordinate all the various
// store parameters on every single new MerkleCRDT creation.
type Factory struct {
	crdts      map[client.CType]*MerkleCRDTFactory
	multistore datastore.MultiStore
}

var (
	// DefaultFactory is instantiated with no stores
	// It is recommended to use this only after you call
	// WithStores(...) so you get a new non-shared instance
	DefaultFactory = NewFactory(nil)
)

// NewFactory returns a newly instantiated factory object with the assigned stores.
// It may be called with all stores set to nil.
func NewFactory(multistore datastore.MultiStore) *Factory {
	return &Factory{
		crdts:      make(map[client.CType]*MerkleCRDTFactory),
		multistore: multistore,
	}
}

// Register creates a new entry in the CRDTs map to register a factory function
// to a MerkleCRDT Type.
func (factory *Factory) Register(t client.CType, fn *MerkleCRDTFactory) error {
	factory.crdts[t] = fn
	return nil
}

// Instance and execute the registered factory function for a given MerkleCRDT type
// supplied with all the current stores (passed in as a datastore.MultiStore object).
func (factory Factory) Instance(
	schemaVersionKey core.CollectionSchemaVersionKey,
	uCh events.UpdateChannel,
	t client.CType,
	key core.DataStoreKey,
	fieldName string,
) (MerkleCRDT, error) {
	// get the factory function for the given MerkleCRDT type
	// and pass in the current factory state as a MultiStore parameter
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}
	return (*fn)(factory, schemaVersionKey, uCh, fieldName)(key), nil
}

// InstanceWithStore executes the registered factory function for the given MerkleCRDT type
// with the additional supplied datastore.MultiStore instead of the saved one on the main Factory.
func (factory Factory) InstanceWithStores(
	store datastore.MultiStore,
	schemaVersionKey core.CollectionSchemaVersionKey,
	uCh events.UpdateChannel,
	t client.CType,
	key core.DataStoreKey,
	fieldName string,
) (MerkleCRDT, error) {
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}

	return (*fn)(store, schemaVersionKey, uCh, fieldName)(key), nil
}

func (factory Factory) getRegisteredFactory(t client.CType) (*MerkleCRDTFactory, error) {
	fn, exists := factory.crdts[t]
	if !exists {
		return nil, ErrFactoryTypeNoExist
	}
	return fn, nil
}

// SetStores sets all the current stores on the Factory in one call.
func (factory *Factory) SetStores(multistore datastore.MultiStore) error {
	factory.multistore = multistore
	return nil
}

// WithStores returns a new instance of the Factory with all the stores set.
func (factory Factory) WithStores(multistore datastore.MultiStore) Factory {
	factory.multistore = multistore
	return factory
}

// Rootstore implements MultiStore.
func (factory Factory) Rootstore() datastore.DSReaderWriter {
	return nil
}

// Data implements datastore.MultiStore and returns the current Datastore.
func (factory Factory) Datastore() datastore.DSReaderWriter {
	if factory.multistore == nil {
		return nil
	}
	return factory.multistore.Datastore()
}

// Head implements datastore.MultiStore and returns the current Headstore.
func (factory Factory) Headstore() datastore.DSReaderWriter {
	if factory.multistore == nil {
		return nil
	}
	return factory.multistore.Headstore()
}

// Peerstore implements datastore.MultiStore and returns the current Peerstore.
func (factory Factory) Peerstore() datastore.DSBatching {
	if factory.multistore == nil {
		return nil
	}
	return factory.multistore.Peerstore()
}

// Head implements datastore.MultiStore and returns the current Headstore.
func (factory Factory) Systemstore() datastore.DSReaderWriter {
	if factory.multistore == nil {
		return nil
	}
	return factory.multistore.Systemstore()
}

// DAGstore implements datastore.MultiStore and returns the current DAGstore.
func (factory Factory) DAGstore() datastore.DAGStore {
	if factory.multistore == nil {
		return nil
	}
	return factory.multistore.DAGstore()
}
