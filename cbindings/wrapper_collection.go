// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include <stdint.h>
#include "defra_structs.h"
extern Result* CollectionCreate(uintptr_t nodePtr, char* json, int isEncrypted,
char* encryptedFields, CollectionOptions options);
extern Result* CollectionDelete(uintptr_t nodePtr, char* docIDStr, char* filterStr, CollectionOptions options);
extern Result* CollectionDescribe(uintptr_t nodePtr, CollectionOptions options);
extern Result* CollectionListDocIDs(uintptr_t nodePtr, CollectionOptions options);
extern Result* CollectionGet(uintptr_t nodePtr, char* docIDStr, int showDeleted, CollectionOptions options);
extern Result* CollectionUpdate(uintptr_t nodePtr, char* docIDStr, char* filterStr,
char* updaterStr, CollectionOptions options);
extern Result* IndexCreate(uintptr_t nodePtr, char* collectionName, char* indexName, char* fieldsStr, int isUnique);
extern Result* IndexList(uintptr_t nodePtr, char* collectionName);
extern Result* IndexDrop(uintptr_t nodePtr, char* collectionName, char* indexName);
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
	def client.CollectionVersion
	w   *CWrapper
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
}

func (c *Collection) CollectionID() string {
	return c.Version().CollectionID
}

func (c *Collection) Create(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	isEncrypted := isEncryptedFromDocCreateOption(opts)
	encryptedFields := encryptedFieldsFromDocCreateOptions(opts)

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(encryptedFields))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	docJSONbytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	cJSON := C.CString(string(docJSONbytes))
	defer C.free(unsafe.Pointer(cJSON))

	res := ConvertAndFreeCResult(C.CollectionCreate(
		C.uintptr_t(c.w.handle),
		cJSON,
		isEncrypted,
		encryptedFields,
		copts,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}

	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...client.DocCreateOption,
) error {
	isEncrypted := isEncryptedFromDocCreateOption(opts)
	encryptedFields := encryptedFieldsFromDocCreateOptions(opts)

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(encryptedFields))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

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
	defer C.free(unsafe.Pointer(cJSON))

	res := ConvertAndFreeCResult(C.CollectionCreate(
		C.uintptr_t(c.w.handle),
		cJSON,
		isEncrypted,
		encryptedFields,
		copts,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
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
	docID := C.CString(doc.ID().String())
	filter := C.CString("")
	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	updater := C.CString(string(document))

	cVersion := C.CString("")
	cCollectionID := C.CString(c.CollectionID())
	cName := C.CString("")
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(updater))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionUpdate(
		C.uintptr_t(c.w.handle),
		docID,
		filter,
		updater,
		copts,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
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
	docIDStr := C.CString(docID.String())
	filter := C.CString("")

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(filter))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionDelete(
		C.uintptr_t(c.w.handle),
		docIDStr,
		filter,
		copts,
	))

	if res.Status != 0 {
		return false, errors.New(res.Error)
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	docIDStr := C.CString(docID.String())
	cShowDeleted := C.int(0)

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionGet(
		C.uintptr_t(c.w.handle),
		docIDStr,
		cShowDeleted,
		copts,
	))

	if res.Status != 0 {
		return false, errors.New(res.Error)
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	docID := C.CString("")
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filterStr := C.CString(string(filterJSON))
	cUpdater := C.CString(updater)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cUpdater))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionUpdate(
		C.uintptr_t(c.w.handle),
		docID,
		filterStr,
		cUpdater,
		copts,
	))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var updateRes client.UpdateResult
	retString := []byte(res.Value)
	if err := json.Unmarshal(retString, &updateRes); err != nil {
		return nil, err
	}
	return &updateRes, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	docID := C.CString("")
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filterStr := C.CString(string(filterJSON))

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionDelete(
		C.uintptr_t(c.w.handle),
		docID,
		filterStr,
		copts,
	))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var deleteRes client.DeleteResult
	retString := []byte(res.Value)
	if err := json.Unmarshal(retString, &deleteRes); err != nil {
		return nil, err
	}
	return &deleteRes, nil
}

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	var cShowDeleted C.int = 0
	if showDeleted {
		cShowDeleted = 1
	}

	docIDStr := C.CString(docID.String())
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.Version().Name)
	cIdentity := identityFromContext(ctx)

	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionGet(
		C.uintptr_t(c.w.handle),
		docIDStr,
		cShowDeleted,
		copts,
	))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	jsonStr := res.Value
	doc, err := client.NewDocWithID(docID, c.Version())
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
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := identityFromContext(ctx)

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.identityPtr = cIdentity
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionListDocIDs(C.uintptr_t(c.w.handle), copts))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	docIDCh := make(chan client.DocIDResult)

	go func() {
		defer close(docIDCh)

		var rawResults []struct {
			DocID string `json:"docID"`
			Error string `json:"error"`
		}

		if err := json.Unmarshal([]byte(res.Value), &rawResults); err != nil {
			docIDCh <- client.DocIDResult{Err: fmt.Errorf("failed to parse docIDs: %w", err)}
			return
		}

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
	name := C.CString(c.def.Name)
	cIndexDescName := C.CString(indexDesc.Name)
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(cIndexDescName))

	orderedFields := make([]string, len(indexDesc.Fields))
	for i, f := range indexDesc.Fields {
		order := "ASC"
		if f.Descending {
			order = "DESC"
		}
		orderedFields[i] = f.Name + ":" + order
	}
	fields := C.CString(strings.Join(orderedFields, ","))
	defer C.free(unsafe.Pointer(fields))

	var cUnique C.int = 0
	if indexDesc.Unique {
		cUnique = 1
	}

	res := ConvertAndFreeCResult(C.IndexCreate(
		C.uintptr_t(c.w.handle),
		name,
		cIndexDescName,
		fields,
		cUnique,
	))

	if res.Status != 0 {
		return client.IndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[client.IndexDescription](res.Value)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	name := C.CString(c.def.Name)
	cIndexName := C.CString(indexName)
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(cIndexName))

	res := ConvertAndFreeCResult(C.IndexDrop(
		C.uintptr_t(c.w.handle),
		name,
		cIndexName,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	name := C.CString(c.def.Name)
	defer C.free(unsafe.Pointer(name))

	res := ConvertAndFreeCResult(C.IndexList(C.uintptr_t(c.w.handle), name))

	if res.Status != 0 {
		return []client.IndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[[]client.IndexDescription](res.Value)
	if err != nil {
		return []client.IndexDescription{}, err
	}
	return retRes, nil
}
