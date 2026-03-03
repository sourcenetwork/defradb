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
extern Result DescribeCollection(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result NewIndex(uintptr_t nodePtr, char* indexName, char* fieldsStr, int isUnique,
CollectionOptions options, uintptr_t identityPtr);
extern Result ListIndexes(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result DeleteIndex(uintptr_t nodePtr, char* indexName, CollectionOptions options, uintptr_t identityPtr);
extern Result NewEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result ListEncryptedIndexes(uintptr_t nodePtr, char* collectionName, uintptr_t identityPtr);
extern Result DeleteEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result TruncateCollection(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern void FreeIdentity(uintptr_t identityPtr);
*/
import "C"

import (
	"context"
	"errors"
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

func (c *Collection) NewIndex(
	ctx context.Context,
	indexDesc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
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
	defer C.FreeIdentity(cIdentity)

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

	res := ConvertAndFreeCResult(C.NewIndex(
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
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
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
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.DeleteIndex(
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
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.IndexDescription, error) {
	cName := C.CString(c.def.Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(C.ListIndexes(C.uintptr_t(c.w.handle), copts, cIdentity))

	if res.Status != 0 {
		return []client.IndexDescription{}, errors.New(res.Error)
	}

	retRes, err := unmarshalResult[[]client.IndexDescription](res.Value)
	if err != nil {
		return []client.IndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) NewEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	name := C.CString(c.def.Name)
	fieldName := C.CString(req.FieldName)
	defer C.free(unsafe.Pointer(name))
	defer C.free(unsafe.Pointer(fieldName))

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.NewEncryptedIndex(
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
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.DeleteEncryptedIndex(
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
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	name := C.CString(c.def.Name)
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(name))
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.ListEncryptedIndexes(C.uintptr_t(c.w.handle), name, cIdentity))

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
	ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	cName := C.CString(c.def.Name)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	res := ConvertAndFreeCResult(
		C.TruncateCollection(
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
