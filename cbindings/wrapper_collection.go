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
extern Result CollectionAdd(uintptr_t nodePtr, char* json, int isEncrypted,
char* encryptedFields, CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionDelete(uintptr_t nodePtr, char* docIDStr, char* filterStr,
CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionDescribe(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionListDocIDs(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionGet(uintptr_t nodePtr, char* docIDStr, int showDeleted,
CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionUpdate(uintptr_t nodePtr, char* docIDStr, char* filterStr,
char* updaterStr, CollectionOptions options, uintptr_t identityPtr);
extern Result IndexAdd(uintptr_t nodePtr, char* indexName, char* fieldsStr, int isUnique,
CollectionOptions options, uintptr_t identityPtr);
extern Result IndexList(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result IndexDelete(uintptr_t nodePtr, char* indexName, CollectionOptions options, uintptr_t identityPtr);
extern Result EncryptedIndexAdd(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result EncryptedIndexList(uintptr_t nodePtr, char* collectionName, uintptr_t identityPtr);
extern Result EncryptedIndexDelete(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result CollectionTruncate(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern void IdentityFree(uintptr_t identityPtr);
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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
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

func (c *Collection) Add(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	addOpts := utils.NewOptions(opts...)
	isEncrypted := 0
	if addOpts.EncryptDoc {
		isEncrypted = 1
	}
	encryptedFields := C.CString("")
	if len(addOpts.EncryptedFields) > 0 {
		encryptedFields = C.CString(strings.Join(addOpts.EncryptedFields, ","))
	}

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := optionToUintptr(addOpts.GetIdentity())
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(encryptedFields))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	docJSONbytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	cJSON := C.CString(string(docJSONbytes))
	defer C.free(unsafe.Pointer(cJSON))

	res := ConvertAndFreeCResult(C.CollectionAdd(
		C.uintptr_t(c.w.handle),
		cJSON,
		C.int(isEncrypted),
		encryptedFields,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}

	doc.Clean()
	return nil
}

func (c *Collection) AddMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	addOpts := utils.NewOptions(opts...)
	isEncrypted := 0
	if addOpts.EncryptDoc {
		isEncrypted = 1
	}
	encryptedFields := C.CString("")
	if len(addOpts.EncryptedFields) > 0 {
		encryptedFields = C.CString(strings.Join(addOpts.EncryptedFields, ","))
	}

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := optionToUintptr(addOpts.GetIdentity())
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(encryptedFields))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
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

	res := ConvertAndFreeCResult(C.CollectionAdd(
		C.uintptr_t(c.w.handle),
		cJSON,
		C.int(isEncrypted),
		encryptedFields,
		copts,
		cIdentity,
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
	opts ...options.Enumerable[options.CollectionUpdateOptions],
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
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(updater))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionUpdate(
		C.uintptr_t(c.w.handle),
		docID,
		filter,
		updater,
		copts,
		cIdentity,
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
	opts ...options.Enumerable[options.CollectionSaveOptions],
) error {
	saveOpt := utils.NewOptions(opts...)
	getOpts := options.CollectionGet().SetShowDeleted(true)
	if saveOpt.Identity.HasValue() {
		getOpts.SetIdentity(saveOpt.Identity.Value())
	}
	_, err := c.Get(ctx, doc.ID(), getOpts)
	if err == nil {
		updateOpts := options.CollectionUpdate()
		if saveOpt.Identity.HasValue() {
			updateOpts.SetIdentity(saveOpt.Identity.Value())
		}
		return c.Update(ctx, doc, updateOpts)
	}
	if strings.Contains(err.Error(), client.ErrDocumentNotFoundOrNotAuthorized.Error()) {
		return c.Add(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionDeleteOptions],
) (bool, error) {
	docIDStr := C.CString(docID.String())
	filter := C.CString("")

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(filter))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionDelete(
		C.uintptr_t(c.w.handle),
		docIDStr,
		filter,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return false, errors.New(res.Error)
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionExistsOptions],
) (bool, error) {
	docIDStr := C.CString(docID.String())
	cShowDeleted := C.int(0)

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionGet(
		C.uintptr_t(c.w.handle),
		docIDStr,
		cShowDeleted,
		copts,
		cIdentity,
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
	opts ...options.Enumerable[options.CollectionUpdateWithFilterOptions],
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
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cUpdater))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionUpdate(
		C.uintptr_t(c.w.handle),
		docID,
		filterStr,
		cUpdater,
		copts,
		cIdentity,
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
	opts ...options.Enumerable[options.CollectionDeleteWithFilterOptions],
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
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionDelete(
		C.uintptr_t(c.w.handle),
		docID,
		filterStr,
		copts,
		cIdentity,
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
	opts ...options.Enumerable[options.CollectionGetOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	var cShowDeleted C.int = 0
	if opt.ShowDeleted {
		cShowDeleted = 1
	}

	docIDStr := C.CString(docID.String())
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.Version().Name)
	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionGet(
		C.uintptr_t(c.w.handle),
		docIDStr,
		cShowDeleted,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	jsonStr := res.Value
	doc, err := client.NewDocWithID(ctx, docID, c.Version())
	if err != nil {
		return nil, err
	}
	err = doc.SetWithJSON(ctx, []byte(jsonStr))
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionGetAllDocIDsOptions],
) (<-chan client.DocIDResult, error) {
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.CollectionListDocIDs(C.uintptr_t(c.w.handle), copts, cIdentity))

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

func (c *Collection) AddIndex(
	ctx context.Context,
	indexDesc client.IndexAddRequest,
	opts ...options.Enumerable[options.CollectionAddIndexOptions],
) (client.IndexDescription, error) {
	cName := C.CString(c.def.Name)
	cIndexDescName := C.CString(indexDesc.Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cIndexDescName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

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

	res := ConvertAndFreeCResult(C.IndexAdd(
		C.uintptr_t(c.w.handle),
		cIndexDescName,
		fields,
		cUnique,
		copts,
		cIdentity,
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

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.CollectionDeleteIndexOptions],
) error {
	cName := C.CString(c.def.Name)
	cIndexName := C.CString(indexName)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cIndexName))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.IndexDelete(
		C.uintptr_t(c.w.handle),
		cIndexName,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionListIndexesOptions],
) ([]client.IndexDescription, error) {
	cName := C.CString(c.def.Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.IndexList(C.uintptr_t(c.w.handle), copts, cIdentity))

	if res.Status != 0 {
		return []client.IndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[[]client.IndexDescription](res.Value)
	if err != nil {
		return []client.IndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) AddEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.AddEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	name := C.CString(c.def.Name)
	fieldName := C.CString(req.FieldName)
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(fieldName))

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.EncryptedIndexAdd(
		C.uintptr_t(c.w.handle),
		name,
		fieldName,
		cIdentity,
	))

	if res.Status != 0 {
		return client.EncryptedIndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[client.EncryptedIndexDescription](res.Value)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	name := C.CString(c.def.Name)
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(cFieldName))

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.EncryptedIndexDelete(
		C.uintptr_t(c.w.handle),
		name,
		cFieldName,
		cIdentity,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	name := C.CString(c.def.Name)
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(name))
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.EncryptedIndexList(C.uintptr_t(c.w.handle), name, cIdentity))

	if res.Status != 0 {
		return []client.EncryptedIndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[[]client.EncryptedIndexDescription](res.Value)
	if err != nil {
		return []client.EncryptedIndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions],
) error {
	cName := C.CString(c.def.Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(
		C.CollectionTruncate(
			C.uintptr_t(c.w.handle),
			copts,
			cIdentity,
		),
	)
	if res.Status != 0 {
		return errors.New(res.Error)
	}

	return nil
}
