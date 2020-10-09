package db

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	badgerds "github.com/ipfs/go-ds-badger"
	logging "github.com/ipfs/go-log"
	mh "github.com/multiformats/go-multihash"
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

const (
	objectMarker = byte(0xff) // @todo: Investigate object marker values
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
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return nil
}

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
	// DocKey verification
	buf, err := doc.Bytes()
	if err != nil {
		return err
	}
	// @todo:  grab the cid Prefix from the DocKey internal CID if available
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}
	// And then feed it some data
	c, err := pref.Sum(buf)
	if err != nil {
		return err
	}
	// fmt.Println(c)
	dockey := key.NewDocKeyV0(c)
	if !dockey.Key.Equal(doc.Key().Key) {
		return errors.Wrap(ErrDocVerification, fmt.Sprintf("Expected %s, got %s", doc.Key().UUID(), dockey.UUID()))
	}

	// check if DocKey exists in DB
	// write object marker
	err = writeObjectMarker(txn.datastore, doc.Key().Instance("v"))
	if err != nil {
		return err
	}
	// write data to DB via MerkleClock/CRDT
	return db.save(txn, doc)
}

// Update an existing document with the new values
// Any field that needs to be removed or cleared
// should call doc.Clear(field) before. Any field that
// is nil/empty that hasn't called Clear will be ignored
func (db *DB) Update(doc *document.Document) error {
	exists, err := db.exists(doc.Key())
	if err != nil {
		return err
	}
	if !exists {
		return ErrDocumentNotFound
	}
	return db.update(doc)
}

// Contract: DB Exists check is already perfomed, and a doc with the given key exists
// Note: Should we CompareAndSet the update, IE: Query the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (db *DB) update(doc *document.Document) error {
	txn, err := db.newTxn(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = db.save(txn, doc)
	if err != nil {
		return err
	}
	return txn.Commit()
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
		return db.update(doc)
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
		if val.IsDirty() {
			err := db.saveValueToMerkleCRDT(txn, doc.Key().ChildString(k), val)
			if err != nil {
				return err
			}
			if val.IsDelete() {
				doc.SetAs(v.Name(), nil, v.Type())
			}
			// set value as clean
			val.Clean()
		}
	}

	return nil
}

// Delete will attempt to delete a document by key
// will return true if a deltion is successful, and
// return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return
// false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given
// DocKey. This includes data, block, and head storage.
func (db *DB) Delete(key key.DocKey) (bool, error) {
	exists, err := db.Exists(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrDocumentNotFound
	}

	// create txn
	txn, err := db.newTxn(false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	// run delete, commit if successful
	deleted, err := db.delete(txn, key)
	if err != nil {
		return false, err
	}
	return deleted, txn.Commit()
}

// at the moment, delete only does data storage delete.
// Dag, and head store will soon follow.
func (db *DB) delete(txn *Txn, key key.DocKey) (bool, error) {
	q := query.Query{
		Prefix:   key.Key.String(),
		KeysOnly: true,
	}
	res, err := txn.datastore.Query(q)

	for e := range res.Next() {
		if e.Error != nil {
			return false, err
		}

		err = txn.datastore.Delete(ds.NewKey(e.Key))
		if err != nil {
			return false, err
		}
	}

	return true, nil
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
		var bytes []byte
		if val.IsDelete() { // empty byte array
			bytes = []byte{}
		} else {
			bytes, err = wval.Bytes()
			if err != nil {
				return err
			}
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
