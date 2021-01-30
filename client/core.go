package client

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

type DB interface {
	// Collections
	CreateCollection(base.CollectionDescription) (Collection, error)
	GetCollection(string) (Collection, error)
	ExecQuery(string) *QueryResult
	SchemaManager() *schema.SchemaManager
	LoadSchema(string) error
	PrintDump()
	GetBlock(c cid.Cid) (blocks.Block, error)
}

type Sequence interface{}

type Txn interface {
	ds.Txn
	core.MultiStore
	Systemstore() core.DSReaderWriter
	IsBatch() bool
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
	UpdateWithFilter(interface{}, interface{}, ...UpdateOpt) (*UpdateResult, error)
	UpdateWithKey(key.DocKey, interface{}, ...UpdateOpt) (*UpdateResult, error)
	UpdateWithKeys([]key.DocKey, interface{}, ...UpdateOpt) (*UpdateResult, error)

	WithTxn(Txn) Collection
}

type UpdateOpt struct{}
type CreateOpt struct{}

type UpdateResult struct {
	Count   int64
	DocKeys []string
}

type QueryResult struct {
	Errors []interface{} `json:"errors,omitempty"`
	Data   interface{}   `json:"data"`
}
