// Copyright 2026 Democratized Data Foundation
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
#include "libdefradb.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (c *Collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

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
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, addOpts.EnableSigning)

	docJSONbytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	cJSON := C.CString(string(docJSONbytes))
	defer C.free(unsafe.Pointer(cJSON))

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddDocument(
		callHandle,
		cJSON,
		C.int(isEncrypted),
		encryptedFields,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	if err := setDocumentIDsFromJSON([]*client.Document{doc}, []byte(res.Value)); err != nil {
		return err
	}

	doc.Clean()

	return nil
}

func (c *Collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
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
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, addOpts.EnableSigning)

	jsonDocs := make([]json.RawMessage, 0, len(docs))
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

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddDocument(
		callHandle,
		cJSON,
		C.int(isEncrypted),
		encryptedFields,
		copts,
		cIdentity,
	))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	if err := setDocumentIDsFromJSON(docs, []byte(res.Value)); err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

// setDocumentIDsFromJSON copies returned IDs onto caller docs.
func setDocumentIDsFromJSON(docs []*client.Document, data []byte) error {
	var docIDs []string
	if err := json.Unmarshal(data, &docIDs); err != nil {
		return err
	}
	if len(docIDs) != len(docs) {
		return client.NewErrUnexpectedType[[]string]("docIDs", docIDs)
	}
	for i, docIDString := range docIDs {
		docID, err := client.NewDocIDFromString(docIDString)
		if err != nil {
			return err
		}
		client.ApplySavedDocumentID(docs[i], docID)
	}
	return nil
}

func setCCollectionSigningOption(copts *C.CollectionOptions, enableSigning immutable.Option[bool]) {
	if !enableSigning.HasValue() {
		return
	}
	if enableSigning.Value() {
		copts.enableSigning = 1
	} else {
		copts.enableSigning = -1
	}
}

func (c *Collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

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
	updateOpts := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(updateOpts.GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filter))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(updater))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, updateOpts.EnableSigning)

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.UpdateDocument(
		callHandle,
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

func (c *Collection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	if !doc.ID().IsValid() {
		return c.AddDocument(ctx, doc, opts...)
	}

	saveOpt := utils.NewOptions(opts...)
	getOpts := options.GetDocument().SetShowDeleted(true)
	if saveOpt.Identity.HasValue() {
		getOpts.SetIdentity(saveOpt.Identity.Value())
	}
	_, err := c.GetDocument(ctx, doc.ID(), getOpts)
	if err == nil {
		updateOpts := options.UpdateDocument()
		if saveOpt.Identity.HasValue() {
			updateOpts.SetIdentity(saveOpt.Identity.Value())
		}
		if saveOpt.EnableSigning.HasValue() {
			updateOpts.SetEnableSigning(saveOpt.EnableSigning.Value())
		}
		return c.UpdateDocument(ctx, doc, updateOpts)
	}
	if strings.Contains(err.Error(), client.ErrDocumentNotFoundOrNotAuthorized.Error()) {
		return c.AddDocument(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	docIDStr := C.CString(docID.String())
	filter := C.CString("")

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	deleteOpts := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(deleteOpts.GetIdentity())
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(filter))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, deleteOpts.EnableSigning)

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.DeleteDocument(
		callHandle,
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

func (c *Collection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	docIDStr := C.CString(docID.String())
	cShowDeleted := C.int(0)

	cVersion := C.CString(c.def.VersionID)
	cCollectionID := C.CString(c.def.CollectionID)
	cName := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(docIDStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.GetDocument(
		callHandle,
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

func (c *Collection) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	docID := C.CString("")
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	filterStr := C.CString(string(filterJSON))
	cUpdater := C.CString(updater)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	updateOpts := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(updateOpts.GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cUpdater))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, updateOpts.EnableSigning)

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.UpdateDocument(
		callHandle,
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

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	docID := C.CString("")
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	filterStr := C.CString(string(filterJSON))

	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString(c.def.Name)
	deleteOpts := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(deleteOpts.GetIdentity())
	defer C.free(unsafe.Pointer(docID))
	defer C.free(unsafe.Pointer(filterStr))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0
	setCCollectionSigningOption(&copts, deleteOpts.EnableSigning)

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.DeleteDocument(
		callHandle,
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

func (c *Collection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

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
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	callHandle := getNodeOrTxnHandle(c.w.handle, ctx)
	res := ConvertAndFreeCResult(C.GetDocument(
		callHandle,
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
