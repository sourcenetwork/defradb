// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !cshared
// +build !cshared

package cwrap

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	def client.CollectionDefinition
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def.Version
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.def.Schema
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
}

func (c *Collection) SchemaRoot() string {
	return c.Schema().Root
}

func (c *Collection) Definition() client.CollectionDefinition {
	return c.def
}

func (c *Collection) Create(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cIsEncrypted := cIsEncryptedFromDocCreateOption(opts)
	cEncryptedFields := cEncryptedFieldsFromDocCreateOptions(opts)
	cName := C.CString(c.def.GetName())
	cVersion := C.CString("")
	cCollectionID := C.CString("")

	docJSONbytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	cJSON := C.CString(string(docJSONbytes))

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = 0

	result := CollectionCreate(cJSON, cIsEncrypted, cEncryptedFields, copts)

	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cJSON))
	defer C.free(unsafe.Pointer(cEncryptedFields))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...client.DocCreateOption,
) error {
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cIsEncrypted := cIsEncryptedFromDocCreateOption(opts)
	cEncryptedFields := cEncryptedFieldsFromDocCreateOptions(opts)
	cName := C.CString(c.Version().Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")

	var jsonDocs []json.RawMessage
	for _, doc := range docs {
		b, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to convert document to JSON: %w", err)
		}
		jsonDocs = append(jsonDocs, b)
	}
	docJSONbytes, err := json.Marshal(jsonDocs)
	if err != nil {
		return err
	}
	cJSON := C.CString(string(docJSONbytes))

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = 0

	result := CollectionCreate(cJSON, cIsEncrypted, cEncryptedFields, copts)

	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cJSON))
	defer C.free(unsafe.Pointer(cEncryptedFields))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	cTxnID := cTxnIDFromContext(ctx)
	cDocID := C.CString(doc.ID().String())
	cIdentity := cIdentityFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString(c.Schema().VersionID)
	cName := C.CString("")
	cFilter := C.CString("")
	var cGetInactive C.int = 0

	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cFilter))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))

	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	cUpdater := C.CString(string(document))

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = cGetInactive
	result := CollectionUpdate(cDocID, cFilter, cUpdater, copts)

	defer C.free(unsafe.Pointer(cUpdater))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	doc.Clean()
	return nil
}

func (c *Collection) Save(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	_, err := c.Get(ctx, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, doc)
	}
	if strings.Contains(err.Error(), client.ErrDocumentNotFoundOrNotAuthorized.Error()) {
		return c.Create(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cVersion := C.CString("")
	cDocID := C.CString(docID.String())
	cFilter := C.CString("")
	cIdentity := cIdentityFromContext(ctx)
	cName := C.CString(c.def.GetName())
	cCollectionID := C.CString("")

	var opts C.CollectionOptions
	opts.tx = cTxnID
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.identity = cIdentity
	opts.getInactive = 0

	result := CollectionDelete(cDocID, cFilter, opts)

	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer freeCResult(result)

	if result.status != 0 {
		return false, errors.New(C.GoString(result.error))
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cDocID := C.CString(docID.String())
	cIdentity := cIdentityFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	var cShowDeleted C.int = 0

	var opts C.CollectionOptions
	opts.tx = cTxnID
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.identity = cIdentity
	opts.getInactive = 0

	result := CollectionGet(cDocID, cShowDeleted, opts)

	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer freeCResult(result)

	if result.status != 0 {
		return false, errors.New(C.GoString(result.error))
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cDocID := C.CString("")
	cIdentity := cIdentityFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.GetName())
	cUpdater := C.CString(updater)
	var cGetInactive C.int = 0

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	cFilter := C.CString(string(filterJSON))

	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cUpdater))
	defer C.free(unsafe.Pointer(cFilter))

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = cGetInactive

	result := CollectionUpdate(cDocID, cFilter, cUpdater, copts)

	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var res client.UpdateResult
	retString := []byte(C.GoString(result.value))
	if err := json.Unmarshal(retString, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cDocID := C.CString("")
	cIdentity := cIdentityFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.GetName())

	var cGetInactive C.int = 0

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	cFilter := C.CString(string(filterJSON))

	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cFilter))

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = cGetInactive

	result := CollectionDelete(cDocID, cFilter, copts)

	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var res client.DeleteResult
	retString := []byte(C.GoString(result.value))
	if err := json.Unmarshal(retString, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cDocID := C.CString(docID.String())
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := cIdentityFromContext(ctx)
	cName := C.CString(c.Version().Name)
	var cShowDeleted C.int = 0
	if showDeleted {
		cShowDeleted = 1
	}

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = 0

	result := CollectionGet(cDocID, cShowDeleted, copts)
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIdentity))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	jsonStr := C.GoString(result.value)
	doc, err := client.NewDocWithID(docID, c.Definition())
	if err != nil {
		return nil, err
	}
	err = doc.SetWithJSON([]byte(jsonStr))
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := cIdentityFromContext(ctx)
	cName := C.CString("")

	var copts C.CollectionOptions
	copts.tx = cTxnID
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identity = cIdentity
	copts.getInactive = 0

	result := CollectionListDocIDs(copts)

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIdentity))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	// We have to convert this now, because the memory for the C string will be freed
	// when the function returns, even though the Go Routine is still working on the string
	resultInGo := C.GoString(result.value)

	docIDCh := make(chan client.DocIDResult)

	go func() {
		defer close(docIDCh)

		// Extract and decode JSON from C-Result
		var rawResults []struct {
			DocID string `json:"docID"`
			Error string `json:"error"`
		}

		if err := json.Unmarshal([]byte(resultInGo), &rawResults); err != nil {
			docIDCh <- client.DocIDResult{Err: fmt.Errorf("failed to parse docIDs: %w", err)}
			return
		}

		// Send each result to the channel
		for _, r := range rawResults {
			docID, err := client.NewDocIDFromString(r.DocID)
			res := client.DocIDResult{
				ID: docID,
			}
			if err != nil {
				res.Err = err
			}
			if r.Error != "" {
				res.Err = errors.New(r.Error)
			}
			docIDCh <- res
		}
	}()

	return docIDCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cName := C.CString(c.def.GetName())
	cIndexName := C.CString(indexDesc.Name)
	var cIsUnique C.int = 0
	if indexDesc.Unique {
		cIsUnique = 1
	}

	// Build the Fields string
	orderedFields := make([]string, len(indexDesc.Fields))
	for i, f := range indexDesc.Fields {
		order := "ASC"
		if f.Descending {
			order = "DESC"
		}
		orderedFields[i] = f.Name + ":" + order
	}
	cFields := C.CString(strings.Join(orderedFields, ","))

	result := IndexCreate(cName, cIndexName, cFields, cIsUnique, cTxnID)

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIndexName))
	defer C.free(unsafe.Pointer(cFields))
	defer freeCResult(result)

	if result.status != 0 {
		return client.IndexDescription{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to client.IndexDescription
	retRes, err := unmarshalResult[client.IndexDescription](result.value)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	cTxnID := cTxnIDFromContext(ctx)
	cName := C.CString(c.def.GetName())
	cIndexName := C.CString(indexName)

	result := IndexDrop(cName, cIndexName, cTxnID)

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIndexName))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	return nil
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cName := C.CString(c.def.GetName())

	result := IndexList(cName, cTxnID)

	defer C.free(unsafe.Pointer(cName))
	defer freeCResult(result)

	if result.status != 0 {
		return []client.IndexDescription{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to []client.IndexDescription
	retRes, err := unmarshalResult[[]client.IndexDescription](result.value)
	if err != nil {
		return []client.IndexDescription{}, err
	}
	return retRes, nil
}
