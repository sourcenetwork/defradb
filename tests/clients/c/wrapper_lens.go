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
	"unsafe"

	"github.com/sourcenetwork/defradb/client"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable/enumerable"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

type LensRegistry struct{}

func (w *LensRegistry) Init(txnSource client.TxnSource) {}

func (w *LensRegistry) SetMigration(ctx context.Context, collectionID string, config model.Lens) error {
	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cCollectionID := C.CString(collectionID)
	cLens := C.CString(string(cfgBytes))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cLens))

	txnID := cTxnIDFromContext(ctx)
	result := LensSetRegistry(cCollectionID, cLens, C.ulonglong(txnID))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *LensRegistry) ReloadLenses(ctx context.Context) error {
	txnID := cTxnIDFromContext(ctx)
	result := LensReload(C.ulonglong(txnID))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *LensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	docs, err := collectEnumerable(src)
	if err != nil {
		return nil, err
	}

	docBytes, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	cCollectionID := C.CString(collectionID)
	cDocs := C.CString(string(docBytes))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cDocs))

	txnID := cTxnIDFromContext(ctx)
	result := LensUp(cCollectionID, cDocs, C.ulonglong(txnID))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(C.GoString(result.value)), &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}

func (w *LensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[map[string]any],
	collectionID string,
) (enumerable.Enumerable[map[string]any], error) {
	docs, err := collectEnumerable(src)
	if err != nil {
		return nil, err
	}

	docBytes, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	cCollectionID := C.CString(collectionID)
	cDocs := C.CString(string(docBytes))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cDocs))

	txnID := cTxnIDFromContext(ctx)
	result := LensDown(cCollectionID, cDocs, C.ulonglong(txnID))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(C.GoString(result.value)), &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}
