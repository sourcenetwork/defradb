package crdt

import (
	"errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
)

var (
	ErrFactoryTypeNoExist = errors.New("No such factory for the given type exists")
)

// MerkleCRDTInitFn intanciates a MerkleCRDT with a given key
type MerkleCRDTInitFn func(ds.Key) MerkleCRDT

// MerkleCRDTFactory instanciates a MerkleCRDTInitFn with a MultiStore
// returns a MerkleCRDTInitFn with all the necessary stores set
type MerkleCRDTFactory func(mstore core.MultiStore) MerkleCRDTInitFn

// Factory is a helper utility for intanciating new MerkleCRDTs.
// It removes some of the overhead of having to coordinate all the various
// store parameters on every single new MerkleCRDT creation
type Factory struct {
	crdts     map[Type]*MerkleCRDTFactory
	datastore store.DSReaderWriter
	headstore store.DSReaderWriter
	dagstore  core.DAGStore
}

var (
	// DefaultFactory is intanciated with no stores
	// It is recommended to use this only after you call
	// WithStores(...) so you get a new non-shared instance
	DefaultFactory = NewFactory(nil, nil, nil)
)

// NewFactory returns a newly instanciated factory object with the assigned stores
// It may be called with all stores set to nil
func NewFactory(datastore, headstore store.DSReaderWriter, dagstore core.DAGStore) *Factory {
	return &Factory{
		crdts:     make(map[Type]*MerkleCRDTFactory),
		datastore: datastore,
		headstore: headstore,
		dagstore:  dagstore,
	}
}

// Register creates a new entry in the crdts map to register a factory function
// to a MerkleCRDT Type.
func (factory *Factory) Register(t Type, fn *MerkleCRDTFactory) error {
	factory.crdts[t] = fn
	return nil
}

// Instance and execute the registered factory function for a given MerkleCRDT type
// supplied with all the current stores (passed in as a core.MultiStore object)
func (factory Factory) Instance(t Type, key ds.Key) (MerkleCRDT, error) {
	// get the factory function for the given MerkleCRDT type
	// and pass in the current factory state as a MultiStore parameter
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}
	return (*fn)(factory)(key), nil
}

// InstanceWithStore executes the registered factory function for the given MerkleCRDT type
// with the additional supplied core.MultiStore instead of the saved one on the main Factory.
func (factory Factory) InstanceWithStore(store core.MultiStore, t Type, key ds.Key) (MerkleCRDT, error) {
	fn, err := factory.getRegisteredFactory(t)
	if err != nil {
		return nil, err
	}

	return (*fn)(store)(key), nil
}

func (factory Factory) getRegisteredFactory(t Type) (*MerkleCRDTFactory, error) {
	fn, exists := factory.crdts[t]
	if !exists {
		return nil, ErrFactoryTypeNoExist
	}
	return fn, nil
}

// SetStores sets all the current stores on the Factory in one call
func (factory *Factory) SetStores(datastore, headstore store.DSReaderWriter, dagstore core.DAGStore) error {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return nil
}

// WithStores returns a new instance of the Factory with all the stores set
func (factory Factory) WithStores(datastore, headstore store.DSReaderWriter, dagstore core.DAGStore) Factory {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return factory
}

// SetDatastore sets the current datastore
func (factory *Factory) SetDatastore(datastore store.DSReaderWriter) error {
	factory.datastore = datastore
	return nil
}

// WithDatastore returns a new copy of the Factory instance with a new Datastore
func (factory Factory) WithDatastore(datastore store.DSReaderWriter) Factory {
	factory.datastore = datastore
	return factory
}

// SetHeadstore sets the current headstore
func (factory *Factory) SetHeadstore(headstore store.DSReaderWriter) error {
	factory.headstore = headstore
	return nil
}

// WithHeadstore returns a new copy of the Factory with a new Headstore
func (factory Factory) WithHeadstore(headstore store.DSReaderWriter) Factory {
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
func (factory Factory) Data() core.DSReaderWriter {
	return factory.datastore
}

// Head implements core.MultiStore and returns the current Headstore
func (factory Factory) Head() core.DSReaderWriter {
	return factory.headstore
}

// Dag implements core.MultiStore and returns the current Dagstore
func (factory Factory) Dag() core.DAGStore {
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
