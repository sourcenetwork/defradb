package db

import (
	"encoding/json"

	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/db/base"
)

var (
	collectionNs = ds.NewKey("/system/collection")
)

type Collection struct {
	db   *DB
	txn  *Txn
	desc base.CollectionDescription
}

// NewCollection returns a pointer to a newly instanciated DB Collection
func (db *DB) NewCollection(desc base.CollectionDescription) (*Collection, error) {
	if desc.Name == "" {
		return nil, errors.New("Collection requires name to not be empty")
	}

	if desc.Schema.IsEmpty() {
		if len(desc.Schema.Fields) == 0 {
			return nil, errors.New("Collection schema has no fields")
		}
		docKeyField := desc.Schema.Fields[0]
		if docKeyField.Kind != base.FieldKind_DocKey || docKeyField.Name != "_key" {
			return nil, errors.New("Collection schema first field must be a DocKey")
		}
		desc.Schema.FieldIDs = make([]uint32, len(desc.Schema.Fields))
		for i, field := range desc.Schema.Fields {
			if field.Name == "" {
				return nil, errors.New("Collection schema field missing Name")
			}
			if field.Kind == base.FieldKind_None {
				return nil, errors.New("Collection schema field missing FieldKind")
			}
			if field.Typ == crdt.NONE_CRDT {
				return nil, errors.New("Collection schema field missing CRDT type")
			}
			desc.Schema.FieldIDs = append(desc.Schema.FieldIDs, uint32(i))
			desc.Schema.Fields[i].ID = uint32(i)
		}
	}

	return &Collection{
		db:   db,
		desc: desc,
	}, nil
}

// CreateCollection
func (db *DB) CreateCollection(desc base.CollectionDescription) (*Collection, error) {
	col, err := db.NewCollection(desc)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(desc)
	if err != nil {
		return nil, err
	}
	key := makeCollectionKey(desc.Name)

	//write the collection metadata to the system store
	err = db.systemstore.Put(key, buf)
	return col, err
}

// GetCollection returns an existing collection within the database
func (db *DB) GetCollection(name string) (*Collection, error) {
	if name == "" {
		return nil, errors.New("Collection name can't be empty")
	}

	key := makeCollectionKey(name)
	buf, err := db.systemstore.Get(key)
	if err != nil {
		return nil, err
	}
	var col *Collection
	err = json.Unmarshal(buf, col)
	return col, err
}

func (c *Collection) ValidDescription() bool {
	return false
}

func (c *Collection) WithTxn(txn *Txn) *Collection {
	return &Collection{
		txn:  txn,
		desc: c.desc,
	}
}

func (c *Collection) Create() {}
func (c *Collection) Update() {}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
func makeCollectionKey(name string) ds.Key {
	return collectionNs.ChildString(name)
}
