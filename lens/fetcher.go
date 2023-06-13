// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"context"
	"reflect"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// todo: The code in here can be significantly simplified with:
// https://github.com/sourcenetwork/defradb/issues/1589

type lensedFetcher struct {
	source   fetcher.Fetcher
	registry client.LensRegistry
	lens     Lens

	txn datastore.Txn

	col *client.CollectionDescription
	// Cache the fieldDescriptions mapped by name to allow for cheaper access within the fetcher loop
	fieldDescriptionsByName map[string]client.FieldDescription

	targetVersionID string
	// We cache the target schema version ID as bytes so we dont have to bother
	// casting it within the fetch loop.
	targetVersionIDBytes []byte
}

var _ fetcher.Fetcher = (*lensedFetcher)(nil)

// NewFetcher returns a new fetcher that will migrate any documents from the given
// source Fetcher as they are are yielded.
func NewFetcher(source fetcher.Fetcher, registry client.LensRegistry) fetcher.Fetcher {
	return &lensedFetcher{
		source:   source,
		registry: registry,
	}
}

func (df *lensedFetcher) Init(
	col *client.CollectionDescription,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docmapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	df.col = col

	df.fieldDescriptionsByName = make(map[string]client.FieldDescription, len(col.Schema.Fields))
	// Add cache the field descriptions in reverse, allowing smaller-index fields to overwrite any later
	// ones.  This should never really happen here, but it ensures the result is consistent with col.GetField
	// which returns the first one it finds with a matching name.
	for i := len(col.Schema.Fields) - 1; i >= 0; i-- {
		field := col.Schema.Fields[i]
		df.fieldDescriptionsByName[field.Name] = field
	}

	df.targetVersionID = col.Schema.VersionID
	df.targetVersionIDBytes = []byte(col.Schema.VersionID)
	return df.source.Init(col, fields, filter, docmapper, reverse, showDeleted)
}

func (df *lensedFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	history, err := getTargetedHistory(ctx, txn, df.registry.Config(), df.col.Schema.SchemaID, df.col.Schema.VersionID)
	if err != nil {
		return err
	}
	df.lens = New(df.registry, df.col.Schema.VersionID, history)
	df.txn = txn

	return df.source.Start(ctx, txn, spans)
}

func (df *lensedFetcher) FetchNext(ctx context.Context) (fetcher.EncodedDocument, error) {
	panic("This function is never called and is dead code.  As this type is internal, panicing is okay for now")
}

func (df *lensedFetcher) FetchNextDecoded(ctx context.Context) (*client.Document, error) {
	doc, err := df.source.FetchNextDecoded(ctx)
	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, nil
	}

	if doc.SchemaVersionID == df.targetVersionID {
		// If the document is already at the target schema version, no migration is required and
		// we can return it early.
		return doc, nil
	}

	sourceLensDoc, err := clientDocToLensDoc(doc)
	if err != nil {
		return nil, err
	}

	err = df.lens.Put(doc.SchemaVersionID, sourceLensDoc)
	if err != nil {
		return nil, err
	}

	hasNext, err := df.lens.Next()
	if err != nil {
		return nil, err
	}
	if !hasNext {
		// The migration decided to not yield a document, so we cycle through the next fetcher doc
		return df.FetchNextDecoded(ctx)
	}

	migratedLensDoc, err := df.lens.Value()
	if err != nil {
		return nil, err
	}

	migratedDoc, err := df.lensDocToClientDoc(migratedLensDoc)
	if err != nil {
		return nil, err
	}

	err = df.updateDataStore(ctx, sourceLensDoc, migratedLensDoc)
	if err != nil {
		return nil, err
	}

	return migratedDoc, nil
}

func (df *lensedFetcher) FetchNextDoc(
	ctx context.Context,
	mapping *core.DocumentMapping,
) ([]byte, core.Doc, error) {
	key, doc, err := df.source.FetchNextDoc(ctx, mapping)
	if err != nil {
		return nil, core.Doc{}, err
	}

	if len(doc.Fields) == 0 {
		return key, doc, nil
	}

	if doc.SchemaVersionID == df.targetVersionID {
		// If the document is already at the target schema version, no migration is required and
		// we can return it early.
		return key, doc, nil
	}

	sourceJson, err := coreDocToLensDoc(mapping, doc)
	if err != nil {
		return nil, core.Doc{}, err
	}
	err = df.lens.Put(doc.SchemaVersionID, sourceJson)
	if err != nil {
		return nil, core.Doc{}, err
	}

	hasNext, err := df.lens.Next()
	if err != nil {
		return nil, core.Doc{}, err
	}
	if !hasNext {
		// The migration decided to not yield a document, so we cycle through the next fetcher doc
		return df.FetchNextDoc(ctx, mapping)
	}

	migratedDocJson, err := df.lens.Value()
	if err != nil {
		return nil, core.Doc{}, err
	}

	migratedDoc, err := df.lensDocToCoreDoc(mapping, migratedDocJson)
	if err != nil {
		return nil, core.Doc{}, err
	}

	return key, migratedDoc, nil
}

func (df *lensedFetcher) Close() error {
	if df.lens != nil {
		df.lens.Reset()
	}
	return df.source.Close()
}

// clientDocToLensDoc converts a client.Document to a LensDoc.
func clientDocToLensDoc(doc *client.Document) (LensDoc, error) {
	docAsMap := map[string]any{}

	for field, fieldValue := range doc.Values() {
		docAsMap[field.Name()] = fieldValue.Value()
	}
	docAsMap[request.KeyFieldName] = doc.Key().String()

	// Note: client.Document does not have a means of flagging as to whether it is
	// deleted or not, and, currently the fetcher does not ever returned deleted items
	// from the function that returs this type.

	return docAsMap, nil
}

// coreDocToLensDoc converts a core.Doc to a LensDoc.
func coreDocToLensDoc(mapping *core.DocumentMapping, doc core.Doc) (LensDoc, error) {
	docAsMap := map[string]any{}

	for fieldIndex, fieldValue := range doc.Fields {
		fieldName, ok := mapping.TryToFindNameFromIndex(fieldIndex)
		if !ok {
			continue
		}
		docAsMap[fieldName] = fieldValue
	}

	docAsMap[request.DeletedFieldName] = doc.Status.IsDeleted()

	return docAsMap, nil
}

// lensDocToCoreDoc converts a LensDoc to a core.Doc.
func (df *lensedFetcher) lensDocToCoreDoc(mapping *core.DocumentMapping, docAsMap LensDoc) (core.Doc, error) {
	doc := mapping.NewDoc()

	for fieldName, fieldByteValue := range docAsMap {
		if fieldName == request.KeyFieldName {
			key, ok := fieldByteValue.(string)
			if !ok {
				return core.Doc{}, core.ErrInvalidKey
			}

			doc.SetKey(key)
			continue
		}

		fieldDesc, fieldFound := df.fieldDescriptionsByName[fieldName]
		if !fieldFound {
			// Note: This can technically happen if a Lens migration returns a field that
			// we do not know about. In which case we have to skip it.
			continue
		}

		fieldValue, err := core.Decode(fieldDesc, fieldByteValue)
		if err != nil {
			return core.Doc{}, err
		}

		index := mapping.FirstIndexOfName(fieldName)
		doc.Fields[index] = fieldValue
	}

	if value, ok := docAsMap[request.DeletedFieldName]; ok {
		if wasDeleted, ok := value.(bool); ok {
			if wasDeleted {
				doc.Status = client.Deleted
			} else {
				doc.Status = client.Active
			}
		}
	}

	doc.SchemaVersionID = df.col.Schema.VersionID

	return doc, nil
}

// lensDocToClientDoc converts a LensDoc to a client.Document.
func (df *lensedFetcher) lensDocToClientDoc(docAsMap LensDoc) (*client.Document, error) {
	key, err := client.NewDocKeyFromString(docAsMap[request.KeyFieldName].(string))
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithKey(key)

	for fieldName, fieldByteValue := range docAsMap {
		if fieldName == request.KeyFieldName {
			continue
		}

		fieldDesc, fieldFound := df.fieldDescriptionsByName[fieldName]
		if !fieldFound {
			// Note: This can technically happen if a Lens migration returns a field that
			// we do not know about. In which case we have to skip it.
			continue
		}

		fieldValue, err := core.Decode(fieldDesc, fieldByteValue)
		if err != nil {
			return nil, err
		}

		err = doc.SetAs(fieldDesc.Name, fieldValue, fieldDesc.Typ)
		if err != nil {
			return nil, err
		}
	}

	doc.SchemaVersionID = df.col.Schema.VersionID

	// Note: client.Document does not have a means of flagging as to whether it is
	// deleted or not, and, currently the fetcher does not ever returned deleted items
	// from the function that returs this type.

	return doc, nil
}

// updateDataStore updates the datastore with the migrated values.
//
// This removes the need to migrate a document everytime it is fetched as the second time around
// the underlying fetcher will return the migrated values cached in the datastore instead of the
// underlying dag store values.
func (df *lensedFetcher) updateDataStore(ctx context.Context, original map[string]any, migrated map[string]any) error {
	modifiedFieldValuesByName := map[string]any{}
	for name, originalValue := range original {
		migratedValue, ok := migrated[name]
		if !ok {
			// If the field is present in the original, and missing from the migrated, it
			// means that a migration has removed it, and we should set it to nil.
			modifiedFieldValuesByName[name] = nil
			continue
		}

		// Note: A deep equals check is required here, as the values may be inline-array slices
		if !reflect.DeepEqual(originalValue, migratedValue) {
			modifiedFieldValuesByName[name] = migratedValue
		}
	}

	for name, migratedValue := range migrated {
		if _, ok := original[name]; !ok {
			// If a field has been added by a migration we need to make sure we
			// preserve it here.
			modifiedFieldValuesByName[name] = migratedValue
			continue
		}
	}

	if len(modifiedFieldValuesByName) > 0 {
		dockey, ok := original[request.KeyFieldName].(string)
		if !ok {
			return core.ErrInvalidKey
		}

		datastoreKeyBase := core.DataStoreKey{
			CollectionID: df.col.IDString(),
			DocKey:       dockey,
			InstanceType: core.ValueKey,
		}

		for fieldName, value := range modifiedFieldValuesByName {
			fieldDesc, ok := df.fieldDescriptionsByName[fieldName]
			if !ok {
				// It may be that the migration has set fields that are unknown to us locally
				// in which case we have to skip them for now.
				continue
			}
			fieldKey := datastoreKeyBase.WithFieldId(fieldDesc.ID.String())

			bytes, err := cbor.Marshal(value)
			if err != nil {
				return err
			}

			err = df.txn.Datastore().Put(ctx, fieldKey.ToDS(), append([]byte{byte(fieldDesc.Typ)}, bytes...))
			if err != nil {
				return err
			}
		}

		versionKey := datastoreKeyBase.WithFieldId(core.DATASTORE_DOC_VERSION_FIELD_ID)
		err := df.txn.Datastore().Put(ctx, versionKey.ToDS(), df.targetVersionIDBytes)
		if err != nil {
			return err
		}
	}

	return nil
}
