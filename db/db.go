package db

import (
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	badgerds "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-log"

	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/store"
)

// DB is the main interface for interacting with the
// DefraDB storage system.
//
type DB struct {
	rootstore ds.Batching // main storage interface

	datastore ds.Batching     // wrapped store for data
	headstore ds.Batching     // wrapped store for heads
	dagstore  *store.DAGStore // wrapped store for dags

	factory *crdt.Factory

	log log.StandardLogger
}

// Options for database
type Options struct {
	Store  string
	Memory MemoryOptions
	Badger BadgerOptions
}

// BadgerOptions for the badger instance of the backing datastore
type BadgerOptions struct {
	Path string
	*badgerds.Options
}

// MemoryOptions for the memory instance of the backing datastore
type MemoryOptions struct {
	Size uint64
}

// NewDB creates a new instance of the DB using the given options
func NewDB(options *Options) (*DB, error) {
	var rootstore ds.Batching
	var err error
	if options.Store == "badger" {
		rootstore, err = badgerds.NewDatastore(options.Badger.Path, options.Badger.Options)
		if err != nil {
			return nil, err
		}
	} else if options.Store == "memory" {
		rootstore = ds.NewMapDatastore()
	}

	datastore := namespace.Wrap(rootstore, ds.NewKey("/db/data"))
	headstore := namespace.Wrap(rootstore, ds.NewKey("/db/heads"))
	dagstore := store.NewDAGStore(namespace.Wrap(rootstore, ds.NewKey("/db/blocks")))
	factory := crdt.DefaultFactory.WithStores(datastore, headstore, dagstore)

	return &DB{
		rootstore: rootstore,
		datastore: datastore,
		headstore: headstore,
		dagstore:  dagstore,
		factory:   &factory,
		log:       log.Logger("defra.db"),
	}, nil
}

/*

db.newTx

*/

// Save a document into the db
// Either by creating a new document
// or by updating an existing one
func (db *DB) Save(doc *document.Document) error {
	// New batch transaction/store
	// Loop through doc values
	//	=> 		instanciate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values

	for k, v := range doc.Fields() {
		val, _ := doc.ValueOfField(v)
		_ = db.commitValueToMerkleCRDT(doc.Key().ChildString(k), val)
	}

	return nil
}

func (db *DB) commitValueToMerkleCRDT(key ds.Key, val document.Value) error {
	switch val.Type() {
	case crdt.LWW_REGISTER:
		wval, ok := val.(document.WriteableValue)
		if !ok {
			return document.ErrValueTypeMismatch
		}
		datatype := db.factory.Instance(crdt.LWW_REGISTER, key).(*crdt.MerkleLWWRegister)
		bytes, err := wval.Bytes()
		if err != nil {
			return err
		}
		return datatype.Set(bytes)
		// data.Set()
	case crdt.OBJECT:
		db.log.Debug("Sub objects not yet supported")
		break
	}
	return nil
}
