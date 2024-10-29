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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// todo: The code in here can be significantly simplified with:
// https://github.com/sourcenetwork/defradb/issues/1589

type lensedFetcher struct {
	source   fetcher.Fetcher
	registry client.LensRegistry
	lens     Lens

	txn datastore.Txn

	col client.Collection

	// Cache the fieldDescriptions mapped by name to allow for cheaper access within the fetcher loop
	fieldDescriptionsByName map[string]client.FieldDefinition

	targetVersionID string

	// If true there are migrations registered for the collection being fetched.
	hasMigrations bool
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

func (f *lensedFetcher) Init(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	col client.Collection,
	fields []client.FieldDefinition,
	filter *mapper.Filter,
	docmapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	f.col = col

	f.fieldDescriptionsByName = make(map[string]client.FieldDefinition, len(col.Schema().Fields))
	// Add cache the field descriptions in reverse, allowing smaller-index fields to overwrite any later
	// ones.  This should never really happen here, but it ensures the result is consistent with col.GetField
	// which returns the first one it finds with a matching name.
	defFields := col.Definition().GetFields()
	for i := len(defFields) - 1; i >= 0; i-- {
		f.fieldDescriptionsByName[defFields[i].Name] = defFields[i]
	}

	history, err := getTargetedCollectionHistory(ctx, txn, f.col.Schema().Root, f.col.Schema().VersionID)
	if err != nil {
		return err
	}
	f.lens = new(ctx, f.registry, f.col.Schema().VersionID, history)
	f.txn = txn

historyLoop:
	for _, historyItem := range history {
		sources := historyItem.collection.CollectionSources()
		for _, source := range sources {
			if source.Transform.HasValue() {
				f.hasMigrations = true
				break historyLoop
			}
		}
	}

	f.targetVersionID = col.Schema().VersionID

	var innerFetcherFields []client.FieldDefinition
	if f.hasMigrations {
		// If there are migrations present, they may require fields that are not otherwise
		// requested.  At the moment this means we need to pass in nil so that the underlying
		// fetcher fetches everything.
		innerFetcherFields = nil
	} else {
		innerFetcherFields = fields
	}
	return f.source.Init(
		ctx,
		identity,
		txn,
		acp,
		col,
		innerFetcherFields,
		filter,
		docmapper,
		reverse,
		showDeleted,
	)
}

func (f *lensedFetcher) Start(ctx context.Context, spans core.Spans) error {
	return f.source.Start(ctx, spans)
}

func (f *lensedFetcher) FetchNext(ctx context.Context) (fetcher.EncodedDocument, fetcher.ExecInfo, error) {
	doc, execInfo, err := f.source.FetchNext(ctx)
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	if doc == nil {
		return nil, execInfo, nil
	}

	if !f.hasMigrations || doc.SchemaVersionID() == f.targetVersionID {
		// If there are no migrations registered for this schema, or if the document is already
		// at the target schema version, no migration is required and we can return it early.
		return doc, execInfo, nil
	}

	sourceLensDoc, err := encodedDocToLensDoc(doc)
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	err = f.lens.Put(doc.SchemaVersionID(), sourceLensDoc)
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	hasNext, err := f.lens.Next()
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}
	if !hasNext {
		// The migration decided to not yield a document, so we cycle through the next fetcher doc
		doc, nextExecInfo, err := f.FetchNext(ctx)
		execInfo.Add(nextExecInfo)
		return doc, execInfo, err
	}

	migratedLensDoc, err := f.lens.Value()
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	migratedDoc, err := f.lensDocToEncodedDoc(migratedLensDoc)
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	err = f.updateDataStore(ctx, sourceLensDoc, migratedLensDoc)
	if err != nil {
		return nil, fetcher.ExecInfo{}, err
	}

	return migratedDoc, execInfo, nil
}

func (f *lensedFetcher) Close() error {
	if f.lens != nil {
		f.lens.Reset()
	}
	return f.source.Close()
}

// encodedDocToLensDoc converts a [fetcher.EncodedDocument] to a LensDoc.
func encodedDocToLensDoc(doc fetcher.EncodedDocument) (LensDoc, error) {
	docAsMap := map[string]any{}

	properties, err := doc.Properties(false)
	if err != nil {
		return nil, err
	}

	for field, fieldValue := range properties {
		docAsMap[field.Name] = fieldValue
	}
	docAsMap[request.DocIDFieldName] = string(doc.ID())

	// Note: client.Document does not have a means of flagging as to whether it is
	// deleted or not, and, currently the fetcher does not ever returned deleted items
	// from the function that returs this type.

	return docAsMap, nil
}

func (f *lensedFetcher) lensDocToEncodedDoc(docAsMap LensDoc) (fetcher.EncodedDocument, error) {
	var key string
	status := client.Active
	properties := map[client.FieldDefinition]any{}

	for fieldName, fieldByteValue := range docAsMap {
		if fieldName == request.DocIDFieldName {
			key = fieldByteValue.(string)
			continue
		}

		if fieldName == request.DeletedFieldName {
			if wasDeleted, ok := fieldByteValue.(bool); ok {
				if wasDeleted {
					status = client.Deleted
				}
			}
			continue
		}

		fieldDesc, fieldFound := f.fieldDescriptionsByName[fieldName]
		if !fieldFound {
			// Note: This can technically happen if a Lens migration returns a field that
			// we do not know about. In which case we have to skip it.
			continue
		}

		fieldValue, err := core.NormalizeFieldValue(fieldDesc, fieldByteValue)
		if err != nil {
			return nil, err
		}

		properties[fieldDesc] = fieldValue
	}

	return &lensEncodedDocument{
		key:             []byte(key),
		schemaVersionID: f.col.Schema().VersionID,
		status:          status,
		properties:      properties,
	}, nil
}

// updateDataStore updates the datastore with the migrated values.
//
// This removes the need to migrate a document everytime it is fetched as the second time around
// the underlying fetcher will return the migrated values cached in the datastore instead of the
// underlying dag store values.
func (f *lensedFetcher) updateDataStore(ctx context.Context, original map[string]any, migrated map[string]any) error {
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
		// Todo: `reflect.DeepEqual` is pretty rubish long-term here and should be replaced
		// with something more defra specific: https://github.com/sourcenetwork/defradb/issues/1606
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

	docID, ok := original[request.DocIDFieldName].(string)
	if !ok {
		return core.ErrInvalidKey
	}

	datastoreKeyBase := keys.DataStoreKey{
		CollectionRootID: f.col.Description().RootID,
		DocID:            docID,
		InstanceType:     keys.ValueKey,
	}

	for fieldName, value := range modifiedFieldValuesByName {
		fieldDesc, ok := f.fieldDescriptionsByName[fieldName]
		if !ok {
			// It may be that the migration has set fields that are unknown to us locally
			// in which case we have to skip them for now.
			continue
		}
		fieldKey := datastoreKeyBase.WithFieldID(fieldDesc.ID.String())

		bytes, err := cbor.Marshal(value)
		if err != nil {
			return err
		}

		err = f.txn.Datastore().Put(ctx, fieldKey.ToDS(), bytes)
		if err != nil {
			return err
		}
	}

	versionKey := datastoreKeyBase.WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID)
	err := f.txn.Datastore().Put(ctx, versionKey.ToDS(), []byte(f.targetVersionID))
	if err != nil {
		return err
	}

	return nil
}

type lensEncodedDocument struct {
	key             []byte
	schemaVersionID string
	status          client.DocumentStatus
	properties      map[client.FieldDefinition]any
}

var _ fetcher.EncodedDocument = (*lensEncodedDocument)(nil)

func (encdoc *lensEncodedDocument) ID() []byte {
	return encdoc.key
}

func (encdoc *lensEncodedDocument) SchemaVersionID() string {
	return encdoc.schemaVersionID
}

func (encdoc *lensEncodedDocument) Status() client.DocumentStatus {
	return encdoc.status
}

func (encdoc *lensEncodedDocument) Properties(onlyFilterProps bool) (map[client.FieldDefinition]any, error) {
	return encdoc.properties, nil
}

func (encdoc *lensEncodedDocument) Reset() {
	encdoc.key = nil
	encdoc.schemaVersionID = ""
	encdoc.status = 0
	encdoc.properties = map[client.FieldDefinition]any{}
}
