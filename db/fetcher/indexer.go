package fetcher

import (
	"bytes"
	"context"
	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"

	"github.com/ipfs/go-datastore/query"
)

type IndexFetcher struct {
	docFetcher         Fetcher
	col                *client.CollectionDescription
	txn                datastore.Txn
	filter             *mapper.Filter
	doc                *encodedDocument
	index              client.IndexDescription
	indexedField       client.FieldDescription
	docFields          []client.FieldDescription
	indexQuery         query.Results
	indexDataStoreKey  core.IndexDataStoreKey
	indexFilterCond    any
	indexQueryProvider filteredIndexQueryProvider
}

var _ Fetcher = (*IndexFetcher)(nil)

func NewIndexFetcher(
	docFetcher Fetcher,
	indexedFieldDesc client.FieldDescription,
	indexDesc client.IndexDescription,
	filterCond any,
) *IndexFetcher {
	return &IndexFetcher{
		docFetcher:      docFetcher,
		indexedField:    indexedFieldDesc,
		index:           indexDesc,
		indexFilterCond: filterCond,
	}
}

type filteredIndexQueryProvider interface {
	Get(context.Context, datastore.Txn) (query.Results, error)
}

type eqIndexQueryProvider struct {
	indexKey  core.IndexDataStoreKey
	filterVal []byte
}

func (i *eqIndexQueryProvider) Get(ctx context.Context, txn datastore.Txn) (query.Results, error) {
	if len(i.indexKey.FieldValues) != 0 {
		return nil, nil
	}

	i.indexKey.FieldValues = [][]byte{i.filterVal}
	return txn.Datastore().Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
	})
}

type gtIndexQueryProvider struct {
	indexKey  core.IndexDataStoreKey
	filterVal []byte
}

type gtIndexCmp struct {
	value []byte
}

func (cmp *gtIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	res := bytes.Compare(indexKey.FieldValues[0], cmp.value)
	return res > 0
}

func (i *gtIndexQueryProvider) Get(ctx context.Context, txn datastore.Txn) (query.Results, error) {
	return txn.Datastore().Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
		Filters: []query.Filter{
			&gtIndexCmp{value: i.filterVal},
		},
	})
}

func (f *IndexFetcher) createFilteredIndexQueryProvider(indexFilterCond any) (filteredIndexQueryProvider, error) {
	condMap, ok := indexFilterCond.(map[string]any)
	if !ok {
		return nil, errors.New("invalid index filter condition")
	}
	var op string
	var filterVal any
	for op, filterVal = range condMap {
		break
	}

	writableValue := client.NewCBORValue(client.LWW_REGISTER, filterVal)

	valueBytes, err := writableValue.Bytes()
	if err != nil {
		return nil, err
	}

	if op == "_eq" {
		return &eqIndexQueryProvider{
			indexKey:  f.indexDataStoreKey,
			filterVal: valueBytes,
		}, nil
	} else if op == "_gt" {
		return &gtIndexQueryProvider{
			indexKey:  f.indexDataStoreKey,
			filterVal: valueBytes,
		}, nil
	}

	return nil, errors.New("invalid index filter condition")
}

func (f *IndexFetcher) Init(
	col *client.CollectionDescription,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	f.col = col
	f.filter = filter
	f.doc = &encodedDocument{}
	f.doc.mapping = docMapper

	f.indexDataStoreKey.CollectionID = f.col.ID
	f.indexDataStoreKey.IndexID = f.index.ID

	for i := range fields {
		if fields[i].Name == f.indexedField.Name {
			f.docFields = append(fields[:i], fields[i+1:]...)
		}
	}

	queryProvider, err := f.createFilteredIndexQueryProvider(f.indexFilterCond)
	if err != nil {
		return err
	}
	f.indexQueryProvider = queryProvider

	return nil
}

func (f *IndexFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	f.txn = txn

	var err error
	f.indexQuery, err = f.indexQueryProvider.Get(ctx, f.txn)
	if err != nil {
		return err
	}
	return nil
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, error) {
	f.doc.Reset()

	res, hasValue := f.indexQuery.NextSync()
	if !hasValue || res.Error != nil {
		return nil, res.Error
	}

	indexKey, err := core.NewIndexDataStoreKey(res.Key)
	if err != nil {
		return nil, err
	}
	property := &encProperty{
		Desc: f.indexedField,
		Raw:  indexKey.FieldValues[0],
	}

	f.doc.key = indexKey.FieldValues[1]
	f.doc.Properties = append(f.doc.Properties, property)

	if f.docFetcher != nil {
		targetKey := base.MakeDocKey(*f.col, string(f.doc.key))
		spans := core.NewSpans(core.NewSpan(targetKey, targetKey.PrefixEnd()))
		err = f.docFetcher.Init(f.col, f.docFields, f.filter, f.doc.mapping, false, false)
		if err != nil {
			return nil, err
		}
		err = f.docFetcher.Start(ctx, f.txn, spans)
		if err != nil {
			return nil, err
		}
		encDoc, err := f.docFetcher.FetchNext(ctx)
		if err != nil {
			return nil, err
		}
		err = f.docFetcher.Close()
		if err != nil {
			return nil, err
		}
		f.doc.MergeProperties(encDoc)
	}
	return f.doc, nil
}

func (f *IndexFetcher) FetchNextDecoded(ctx context.Context) (*client.Document, error) {
	encDoc, err := f.FetchNext(ctx)
	if err != nil {
		return nil, err
	}
	if encDoc == nil {
		return nil, nil
	}

	decodedDoc, err := encDoc.Decode()
	if err != nil {
		return nil, err
	}

	return decodedDoc, nil
}

func (f *IndexFetcher) FetchNextDoc(ctx context.Context, mapping *core.DocumentMapping) ([]byte, core.Doc, error) {
	encDoc, err := f.FetchNext(ctx)
	if err != nil {
		return nil, core.Doc{}, err
	}
	if encDoc == nil {
		return nil, core.Doc{}, nil
	}

	doc, err := encDoc.DecodeToDoc()
	if err != nil {
		return nil, core.Doc{}, err
	}
	doc.Status = client.Active
	return encDoc.Key(), doc, err
}

func (f *IndexFetcher) Close() error {
	if f.indexQuery != nil {
		return f.indexQuery.Close()
	}
	return nil
}
