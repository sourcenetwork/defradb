package fetcher

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

type IndexFetcher struct {
	nonIndexFetcher Fetcher
	col             *client.CollectionDescription
	txn             datastore.Txn
	indexedFields   []client.FieldDescription
	filter          *mapper.Filter
	reverse         bool
	showDeleted     bool
	doc             *encodedDocument
	index           client.IndexDescription
	indexCond       any
	indexedField    client.FieldDescription
	didReturn       bool
}

var _ Fetcher = (*IndexFetcher)(nil)

func NewIndexFetcher(
	nonIndexFetcher Fetcher,
	indexedFields []client.FieldDescription,
) *IndexFetcher {
	return &IndexFetcher{
		nonIndexFetcher: nonIndexFetcher,
		indexedFields:   indexedFields,
	}
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
	fields = f.filterOutIndexedFields(fields)
	f.doc = &encodedDocument{}
	f.doc.mapping = docMapper
	f.reverse = reverse
	f.showDeleted = showDeleted

	colIndexes := f.col.Indexes
	for fieldName, cond := range f.filter.ExternalConditions {
		for i := range colIndexes {
			if fieldName == colIndexes[i].Fields[0].Name {
				f.index = colIndexes[i]
				f.indexCond = cond
				for j := range f.indexedFields {
					if f.indexedFields[j].Name == fieldName {
						f.indexedField = f.indexedFields[j]
						break
					}
				}
			}
		}
	}

	if len(fields) == 0 {
		f.nonIndexFetcher = nil
		return nil
	}
	return f.nonIndexFetcher.Init(col, fields, filter, docMapper, reverse, showDeleted)
}

func (f *IndexFetcher) filterOutIndexedFields(
	fields []client.FieldDescription,
) []client.FieldDescription {
	fieldLen := len(fields)
	indexedLen := len(f.indexedFields)
	for i := 0; i < indexedLen; {
		isFound := false
		for j := 0; j < fieldLen; j++ {
			if fields[j].Name == f.indexedFields[i].Name {
				isFound = true
				fieldLen--
				fields[j] = fields[fieldLen]
				i++
				break
			}
		}
		if !isFound {
			indexedLen--
			f.indexedFields[i] = f.indexedFields[indexedLen]
		}
		i++
	}
	fields = fields[:fieldLen]
	f.indexedFields = f.indexedFields[:indexedLen]
	return fields
}

func (f *IndexFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	if !f.canUseIndex() {
		return f.nonIndexFetcher.Start(ctx, txn, spans)
	}
	f.txn = txn
	return nil
}

func (f *IndexFetcher) canUseIndex() bool {
	return f.index.ID != 0
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, error) {
	if !f.canUseIndex() {
		return f.nonIndexFetcher.FetchNext(ctx)
	}

	if f.didReturn {
		return nil, nil
	}

	f.doc.Reset()

	condMap, ok := f.indexCond.(map[string]any)
	if !ok {
		return nil, nil
	}
	var op string
	var filterVal any
	for op, filterVal = range condMap {
		break
	}

	if op != "_eq" {
		return nil, nil
	}

	filterStrVal, ok := filterVal.(string)
	if !ok {
		return nil, nil
	}

	writableValue := client.NewCBORValue(client.LWW_REGISTER, filterStrVal)
	//indexDataStoreKey.FieldValues = [][]byte{fieldValue, []byte(doc.Key().String())}

	valBytes, err := writableValue.Bytes()
	if err != nil {
		return nil, err
	}

	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = f.col.ID
	indexDataStoreKey.IndexID = f.index.ID
	indexDataStoreKey.FieldValues = [][]byte{valBytes}

	keys, err := datastore.FetchKeysForPrefix(ctx, indexDataStoreKey.ToString(), f.txn.Datastore())
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		indexKey, err := core.NewIndexDataStoreKey(key.String())
		if err != nil {
			return nil, err
		}
		property := &encProperty{
			Desc: f.indexedField,
			Raw:  append([]byte{byte(client.LWW_REGISTER)}, valBytes...),
		}

		f.doc.key = indexKey.FieldValues[1]
		f.doc.Properties = append(f.doc.Properties, property)

		if f.nonIndexFetcher != nil {
			/*targetKey := base.MakeDocKey(*f.col, string(f.doc.key))
			spans := core.NewSpans(core.NewSpan(targetKey, targetKey.PrefixEnd()))
			err = f.nonIndexFetcher.Start(ctx, f.txn, spans)
			if err != nil {
				return nil, err
			}
			encDoc, err := f.nonIndexFetcher.FetchNext(ctx)
			if err != nil {
				return nil, err
			}
			f.doc.MergeProperties(encDoc)*/
		}
		f.didReturn = true
		return f.doc, nil
	}

	return nil, nil
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
	if !f.canUseIndex() {
		return f.nonIndexFetcher.Close()
	}
	return nil
}
