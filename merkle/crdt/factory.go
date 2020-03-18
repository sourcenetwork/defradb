package crdt

import (
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/store"
)

type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
type MerkleCRDTFactory func(mstore core.MultiStore) MerkleCRDTInitFn

// Factory is a helper utility for intanciating new MerkleCRDTs.
// It removes some of the overhead of having to coordinate all the various
// store parameters on every single new MerkleCRDT creation
type Factory struct {
	crdts     map[Type]MerkleCRDTFactory
	datastore ds.Datastore
	headstore ds.Datastore
	dagstore  *store.DAGStore
	log       logging.StandardLogger
}

var (
	// DefaultFactory is intanciated with no stores
	// It is recommended to use this only after you call
	// WithStores(...) so you get a new non-shared instance
	DefaultFactory = NewFactory(nil, nil, nil)
)

// NewFactory returns a newly instanciated factory object with the assigned stores
// It may be called with all stores set to nil
func NewFactory(datastore, headstore ds.Datastore, dagstore *store.DAGStore) *Factory {
	return &Factory{
		crdts:     make(map[Type]MerkleCRDTFactory),
		datastore: datastore,
		headstore: headstore,
		dagstore:  dagstore,
	}
}

// Register creates a new entry in the crdts map to register a factory function
// to a MerkleCRDT Type.
func (factory *Factory) Register(t Type, fn MerkleCRDTFactory) error {
	factory.crdts[t] = fn
	return nil
}

// Get and execute the registered factory function for a given MerkleCRDT type
// supplied with all the current stores (passed in as a core.MultiStore object)
func (factory Factory) Get(t Type, key ds.Key) MerkleCRDT {
	// get the factory function for the given MerkleCRDT type
	// and pass in the current factory state as a MultiStore parameter
	fn := factory.crdts[t]
	return fn(factory)(key)
}

// SetStores sets all the current stores on the Factory in one call
func (factory *Factory) SetStores(datastore, headstore ds.Datastore, dagstore *store.DAGStore) error {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return nil
}

// WithStores returns a new instance of the Factory with all the stores set
func (factory Factory) WithStores(datastore, headstore ds.Datastore, dagstore *store.DAGStore) Factory {
	factory.datastore = datastore
	factory.headstore = headstore
	factory.dagstore = dagstore
	return factory
}

// SetDatastore sets the current datastore
func (factory *Factory) SetDatastore(datastore ds.Datastore) error {
	factory.datastore = datastore
	return nil
}

// WithDatastore returns a new Factory instance with a new Datastore
func (factory Factory) WithDatastore(datastore ds.Datastore) Factory {
	factory.datastore = datastore
	return factory
}

// SetHeadstore sets the current headstore
func (factory *Factory) SetHeadstore(headstore ds.Datastore) error {
	factory.headstore = headstore
	return nil
}

// WithHeadstore returns a new Factory with a new Headstore
func (factory Factory) WithHeadstore(headstore ds.Datastore) Factory {
	factory.headstore = headstore
	return factory
}

// SetDagstore sets the current dagstore
func (factory *Factory) SetDagstore(dagstore *store.DAGStore) error {
	factory.dagstore = dagstore
	return nil
}

// WithDagstore returns a new Factory with a new Dagstore
func (factory Factory) WithDagstore(dagstore *store.DAGStore) Factory {
	factory.dagstore = dagstore
	return factory
}

// Data implements core.MultiStore and returns the current Datastore
func (factory Factory) Data() ds.Datastore {
	return factory.datastore
}

// Head implements core.MultiStore and returns the current Headstore
func (factory Factory) Head() ds.Datastore {
	return factory.headstore
}

// Dag implements core.MultiStore and returns the current Dagstore
func (factory Factory) Dag() *store.DAGStore {
	return factory.dagstore
}

func (factory Factory) Log() logging.StandardLogger {
	return factory.log
}

/* API Example

factory.RegisterCRDT(LWW_REGISTER, store, InitFn)
lww := factory.Get(LWW_REGISTER)(ds.NewKey("MeyKey")).(*MerkleLWWRegister)

*/
