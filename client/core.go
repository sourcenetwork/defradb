package client

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"

	"github.com/sourcenetwork/defradb/query/graphql/schema"

	ds "github.com/ipfs/go-datastore"
)

type DB interface {
	// Collections
	CreateCollection(base.CollectionDescription) (Collection, error)
	GetCollection(string) (Collection, error)

	// GetSequence(string) (Sequence, error)
	SchemaManager() *schema.SchemaManager
}

type Sequence interface{}

type Txn interface {
	ds.Txn
	core.MultiStore
	Systemstore() core.DSReaderWriter

	// All DB actions are accessible in a transaction
	//
}

type Collection interface {
	Description() base.CollectionDescription
	Name() string
	Schema() base.SchemaDescription
	ID() uint32

	Indexes() []base.IndexDescription
	PrimaryIndex() base.IndexDescription
	Index(uint32) (base.IndexDescription, error)
	CreateIndex(base.IndexDescription) error

	Create(*document.Document) error
	CreateMany([]*document.Document) error
	Update(*document.Document) error
	Save(*document.Document) error
	Delete(key.DocKey) (bool, error)
	Exists(key.DocKey) (bool, error)

	UpdateWith(interface{}, interface{}, ...UpdateOpt) error
	UpdateWithFilter(interface{}, interface{}, ...UpdateOpt) error
	UpdateWithKey(key.DocKey, interface{}, ...UpdateOpt) error
	UpdateWithKeys([]key.DocKey, interface{}, ...UpdateOpt) error

	WithTxn(Txn) Collection
}

type UpdateOpt struct{}
type CreateOpt struct{}
