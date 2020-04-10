package db

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document/key"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
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

	datastore      ds.Batching // wrapped store for data
	dsKeyTransform ktds.KeyTransform

	headstore      ds.Batching // wrapped store for heads
	hsKeyTransform ktds.KeyTransform

	dagstore        core.DAGStore // wrapped store for dags
	dagKeyTransform ktds.KeyTransform

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
	blockstore := namespace.Wrap(rootstore, ds.NewKey("/db/blocks"))
	dagstore := store.NewDAGStore(blockstore)
	crdtFactory := crdt.DefaultFactory.WithStores(datastore, headstore, dagstore)

	return &DB{
		rootstore: rootstore,

		datastore:      datastore,
		dsKeyTransform: datastore.KeyTransform,

		headstore:      headstore,
		hsKeyTransform: headstore.KeyTransform,

		dagstore:        dagstore,
		dagKeyTransform: blockstore.KeyTransform,

		crdtFactory: &crdtFactory,
		log:         log,
	}, nil
}

/*

db.newTx

*/

// Create a new document
// Will verify the DocKey/CID to ensure that the new document is correctly formatted.
func (db *DB) Create(doc *document.Document) error {
	txn, err := db.newTxn(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.create(txn, doc)
	if err != nil {
		return err
	}
	return txn.Commit()
}

// CreateMany creates a collection of documents at once.
// Will verify the DocKey/CID to ensure that the new documents are correctly formatted.
func (db *DB) CreateMany(docs []*document.Document) error {
	txn, err := db.newTxn(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	for _, doc := range docs {
		err = db.create(txn, doc)
		if err != nil {
			return err
		}
	}
	return txn.Commit()
}

func (db *DB) create(txn *Txn, doc *document.Document) error {
	// add DocKey verification
	err := writeObjectMarker(txn.datastore, doc.Key().Instance("v"))
	if err != nil {
		return err
	}
	return db.save(txn, doc)
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

func (db *DB) save(txn *Txn, doc *document.Document) error {
	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instanciate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	for k, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		err := db.saveValueToMerkleCRDT(txn, doc.Key().ChildString(k), val)
		if err != nil {
			return err
		}
	}

	return nil
}

// Exists checks if a given document exists with supplied DocKey
func (db *DB) Exists(key key.DocKey) (bool, error) {
	return db.exists(key)
}

// check if a document exists with the given key
func (db *DB) exists(key key.DocKey) (bool, error) {
	return db.datastore.Has(key.Key.Instance("v"))
}

func (db *DB) saveValueToMerkleCRDT(txn *Txn, key ds.Key, val document.Value) error {
	switch val.Type() {
	case crdt.LWW_REGISTER:
		wval, ok := val.(document.WriteableValue)
		if !ok {
			return document.ErrValueTypeMismatch
		}
		datatype, err := db.crdtFactory.InstanceWithStores(txn, crdt.LWW_REGISTER, key)
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

func writeObjectMarker(store ds.Write, key ds.Key) error {
	if key.Name() != "v" {
		key = key.Instance("v")
	}
	return store.Put(key, []byte{objectMarker})
}

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
