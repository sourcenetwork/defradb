package db

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document/key"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	badgerds "github.com/ipfs/go-ds-badger"
	logging "github.com/ipfs/go-log"

	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/store"
)

const (
	objectMarker = byte(0xff) // @todo: Investigate object marker values
)

// DB is the main interface for interacting with the
// DefraDB storage system.
//
type DB struct {
	rootstore ds.Batching // main storage interface

	datastore ds.Batching   // wrapped store for data
	headstore ds.Batching   // wrapped store for heads
	dagstore  core.DAGStore // wrapped store for dags

	crdtFactory *crdt.Factory

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
	datastore := namespace.Wrap(rootstore, ds.NewKey("/db/data"))
	headstore := namespace.Wrap(rootstore, ds.NewKey("/db/heads"))
	// only use /db namespace since internal blockstore adds /blocks namespace
	// otherwise we'd set the full namespace to /db/blocks to match the above two stores
	dagstore := store.NewDAGStore(namespace.Wrap(rootstore, ds.NewKey("/db")))
	crdtFactory := crdt.DefaultFactory.WithStores(datastore, headstore, dagstore)

	return &DB{
		rootstore:   rootstore,
		datastore:   datastore,
		headstore:   headstore,
		dagstore:    dagstore,
		crdtFactory: &crdtFactory,
		log:         log,
	}, nil
}

/*

db.newTx

*/

// Create a new document
// Will verify the DocKey/CID to ensure that the new document is correctly
// formatted.
func (db *DB) Create(doc *document.Document) error {
	return db.save(doc)
}

// Update an existing document with the new values
// Any field that needs to be removed or cleared
// should call doc.Clear(field) before. Any field that
// is nil/empty that hasn't called Clear will be ignored
func (db *DB) Update(doc *document.Document) error {
	return nil
}

// Save a document into the db
// Either by creating a new document or by updating an existing one
func (db *DB) Save(doc *document.Document) error {
	// Check if document already exists with key
	exists, err := db.exists(doc.Key())
	if err != nil {
		return err
	}

	if exists {
		return db.Update(doc)
	}
	return db.Create(doc)
}

func (db *DB) save(doc *document.Document) error {
	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instanciate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values

	db.writeObjectMarker(db.datastore, doc.Key().Instance("v"))
	for k, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		err := db.saveValueToMerkleCRDT(doc.Key().ChildString(k), val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) saveValueToMerkleCRDT(key ds.Key, val document.Value) error {
	switch val.Type() {
	case crdt.LWW_REGISTER:
		wval, ok := val.(document.WriteableValue)
		if !ok {
			return document.ErrValueTypeMismatch
		}
		datatype, err := db.crdtFactory.Instance(crdt.LWW_REGISTER, key)
		if err != nil {
			return err
		}
		lwwreg := datatype.(*crdt.MerkleLWWRegister)
		bytes, err := wval.Bytes()
		if err != nil {
			return err
		}
		return lwwreg.Set(bytes)
	case crdt.OBJECT:
		// db.writeObjectMarker(db.datastore, subdoc.Instance("v"))
		db.log.Debug("Sub objects not yet supported")
		break
	}
	return nil
}

// func (db *DB) newTxn(readonly bool) (Txn, error) {
// 	var txn Txn
// 	var err error
// 	txnStore, ok := db.rootstore.(ds.TxnDatastore)
// 	if ok { // we support transactions
// 		dstxn, err = txnStore.NewTransaction(readonly)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return txnToDsShim(txn), nil
// 	} else { // no txn

// 	}
// }

func (db *DB) writeObjectMarker(store ds.Write, key ds.Key) error {
	if key.Name() != "v" {
		key = key.Instance("v")
	}
	return store.Put(key, []byte{objectMarker})
}

// check if a document exists with the given key
func (db *DB) exists(key key.DocKey) (bool, error) {
	return db.datastore.Has(key.Key.Instance("v"))
}

// type basicMultiStore struct {
// 	datastore ds.Datastore
// 	headstore ds.Datastore
// 	dagstore  core.DAGStore
// }

// func (db *DB) stores() core.MultiStore {

// }

func (db *DB) printDebugDB() {
	printStore(db.rootstore)
}

func printStore(store core.DSReaderWriter) {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
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
