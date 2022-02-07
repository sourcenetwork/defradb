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
	"errors"

	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"

	// "github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
)

var (
	ErrFactoryTypeNoExist = errors.New("No such factory for the given type exists")
)

// MerkleCRDTInitFn instantiates a MerkleCRDT with a given key
type MerkleCRDTInitFn func(ds.Key) MerkleCRDT

// MerkleCRDTFactory instantiates a MerkleCRDTInitFn with a MultiStore
// returns a MerkleCRDTInitFn with all the necessary stores set
type MerkleCRDTFactory func(mstore core.MultiStore, schemaID string, bs corenet.Broadcaster) MerkleCRDTInitFn

// Factory is a helper utility for instantiating new MerkleCRDTs.
// It removes some of the overhead of having to coordinate all the various
// store parameters on every single new MerkleCRDT creation
type Factory struct {
	crdts     map[core.CType]*MerkleCRDTFactory
	datastore core.DSReaderWriter
	headstore core.DSReaderWriter
	dagstore  core.DAGStore
}

var (
	// DefaultFactory is instantiated with no stores
	// It is recommended to use this only after you call
	// WithStores(...) so you get a new non-shared instance
	DefaultFactory = NewFactory(nil, nil, nil)
)

// NewFactory returns a newly instanciated factory object with the assigned stores
// It may be called with all stores set to nil
func NewFactory(datastore, headstore core.DSReaderWriter, dagstore core.DAGStore) *Factory {
	return &Factory{
		crdts:     make(map[core.CType]*MerkleCRDTFactory),
		datastore: datastore,
		headstore: headstore,
		dagstore:  dagstore,
	}
}

// Register creates a new entry in the CRDTs map to register a factory function
// to a MerkleCRDT Type.
func (factory *Factory) Register(t core.CType, fn *MerkleCRDTFactory) error {
	factory.crdts[t] = fn
	return nil
}

// Instance and execute the registered factory function for a given MerkleCRDT type
// supplied with all the current stores (passed in as a core.MultiStore object)
func (factory Factory) Instance(schemaID string, bs corenet.Broadcaster, t core.CType, key ds.Key) (MerkleCRDT, error) {
	// get the factory function for the given MerkleCRDT type
	// and pass in the current factory state as a MultiStore parameter
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}
	return (*fn)(factory, schemaID, bs)(key), nil
}

// InstanceWithStore executes the registered factory function for the given MerkleCRDT type
// with the additional supplied core.MultiStore instead of the saved one on the main Factory.
func (factory Factory) InstanceWithStores(store core.MultiStore, schemaID string, bs corenet.Broadcaster, t core.CType, key ds.Key) (MerkleCRDT, error) {
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}

	return (*fn)(store, schemaID, bs)(key), nil
}

func (factory Factory) getRegisteredFactory(t core.CType) (*MerkleCRDTFactory, error) {
	fn, exists := factory.crdts[t]
	if !exists {
		return nil, ErrFactoryTypeNoExist
	}
	return fn, nil
}

// SetStores sets all the current stores on the Factory in one call
func (factory *Factory) SetStores(datastore, headstore core.DSReaderWriter, dagstore core.DAGStore) error {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return nil
}

// WithStores returns a new instance of the Factory with all the stores set
func (factory Factory) WithStores(datastore, headstore core.DSReaderWriter, dagstore core.DAGStore) Factory {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return factory
}

// SetDatastore sets the current datastore
func (factory *Factory) SetDatastore(datastore core.DSReaderWriter) error {
	factory.datastore = datastore
	return nil
}

// WithDatastore returns a new copy of the Factory instance with a new Datastore
func (factory Factory) WithDatastore(datastore core.DSReaderWriter) Factory {
	factory.datastore = datastore
	return factory
}

// SetHeadstore sets the current headstore
func (factory *Factory) SetHeadstore(headstore core.DSReaderWriter) error {
	factory.headstore = headstore
	return nil
}

// WithHeadstore returns a new copy of the Factory with a new Headstore
func (factory Factory) WithHeadstore(headstore core.DSReaderWriter) Factory {
	factory.headstore = headstore
	return factory
}

// SetDagstore sets the current dagstore
func (factory *Factory) SetDagstore(dagstore core.DAGStore) error {
	factory.dagstore = dagstore
	return nil
}

// WithDagstore returns a new copy of the Factory with a new Dagstore
func (factory Factory) WithDagstore(dagstore core.DAGStore) Factory {
	factory.dagstore = dagstore
	return factory
}

// Rootstore impements MultiStore
func (factory Factory) Rootstore() core.DSReaderWriter {
	return nil
}

// SetLogger sets the current logger
// func (factory *Factory) SetLogger(l logging.StandardLogger) error {
// 	factory.log = l
// 	return nil
// }

// WithLogger returns a new copy of the Factory with a new logger
// func (factory Factory) WithLogger(l logging.StandardLogger) Factory {
// 	factory.log = l
// 	return factory
// }

// Data implements core.MultiStore and returns the current Datastore
func (factory Factory) Datastore() core.DSReaderWriter {
	return factory.datastore
}

// Head implements core.MultiStore and returns the current Headstore
func (factory Factory) Headstore() core.DSReaderWriter {
	return factory.headstore
}

// Dag implements core.MultiStore and returns the current Dagstore
func (factory Factory) DAGstore() core.DAGStore {
	return factory.dagstore
}

// Log implements core.MultiStore
// @todo Look into abstracting Log from MutliStore
// func (factory Factory) Log() logging.StandardLogger {
// 	return factory.log
// }

/* API Example

factory.RegisterCRDT(LWW_REGISTER, store, InitFn)
lww := factory.Instance(LWW_REGISTER, key).(*MerkleLWWRegister)

*/
