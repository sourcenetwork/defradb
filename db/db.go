package db

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	badgerds "github.com/ipfs/go-ds-badger"
	logging "github.com/ipfs/go-log"
)

var (
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key
	ErrDocVerification = errors.New("The document verificatioin failed")

	// Individual Store Keys
	rootStoreKey   = ds.NewKey("/db")
	systemStoreKey = rootStoreKey.ChildString("/system")
	dataStoreKey   = rootStoreKey.ChildString("/data")
	headStoreKey   = rootStoreKey.ChildString("/heads")
	blockStoreKey  = rootStoreKey.ChildString("/blocks")
)

// DB is the main interface for interacting with the
// DefraDB storage system.
//
type DB struct {
	glock sync.RWMutex

	rootstore ds.Batching // main storage interface

	systemstore    ds.Batching // wrapped store for system data
	ssKeyTransform ktds.KeyTransform

	datastore      ds.Batching // wrapped store for data
	dsKeyTransform ktds.KeyTransform

	headstore      ds.Batching // wrapped store for heads
	hsKeyTransform ktds.KeyTransform

	dagstore        core.DAGStore // wrapped store for dags
	dagKeyTransform ktds.KeyTransform

	crdtFactory *crdt.Factory

	// indicates if this references an initalized database
	initialized bool

	log logging.StandardLogger
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

	log := logging.Logger("defradb")
	systemstore := namespace.Wrap(rootstore, systemStoreKey)
	datastore := namespace.Wrap(rootstore, dataStoreKey)
	headstore := namespace.Wrap(rootstore, headStoreKey)
	blockstore := namespace.Wrap(rootstore, blockStoreKey)
	dagstore := store.NewDAGStore(blockstore)
	crdtFactory := crdt.DefaultFactory.WithStores(datastore, headstore, dagstore)

	db := &DB{
		rootstore: rootstore,

		systemstore:    systemstore,
		ssKeyTransform: systemstore.KeyTransform,

		datastore:      datastore,
		dsKeyTransform: datastore.KeyTransform,

		headstore:      headstore,
		hsKeyTransform: headstore.KeyTransform,

		dagstore:        dagstore,
		dagKeyTransform: blockstore.KeyTransform,

		crdtFactory: &crdtFactory,
		log:         log,
	}

	err = db.Initialize()
	return db, err
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters
func (db *DB) Initialize() error {
	db.glock.Lock()
	defer db.glock.Unlock()

	if db.initialized { // skip
		return nil
	}

	exists, err := db.systemstore.Has(ds.NewKey("init"))
	if err != nil && err != ds.ErrNotFound {
		return err
	}
	if exists {
		return nil
	}

	//init meta data
	// collection sequence
	_, err = db.getSequence("collection")
	if err != nil {
		return err
	}

	err = db.systemstore.Put(ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) printDebugDB() {
	printStore(db.rootstore)
}

func printStore(store core.DSReaderWriter) {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(q)
	defer results.Close()
	if err != nil {
		panic(err)
	}

	for r := range results.Next() {
		fmt.Println(r.Key, ": ", r.Value)
	}
}
