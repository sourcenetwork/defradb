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
extern Result* LensDown(uintptr_t nodePtr, char* collectionID, char* documents);
extern Result* LensUp(uintptr_t nodePtr, char* collectionID, char* documents);
extern Result* LensReload(uintptr_t nodePtr);
extern Result* LensSetRegistry(uintptr_t nodePtr, char* collectionID, char* cfg);
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

var _ client.LensRegistry = (*LensRegistry)(nil)

type LensRegistry struct {
	*CWrapper
}

func (w *LensRegistry) Init(txnSource client.TxnSource) {
}

func (w *LensRegistry) SetMigration(ctx context.Context, collectionID string, config model.Lens) error {
	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	lens := C.CString(string(cfgBytes))
	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(lens))
	defer C.free(unsafe.Pointer(cCollectionID))

	res := ConvertAndFreeCResult(C.LensSetRegistry(C.uintptr_t(w.handle), cCollectionID, lens))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *LensRegistry) ReloadLenses(ctx context.Context) error {
	res := ConvertAndFreeCResult(C.LensReload(C.uintptr_t(w.handle)))

	if res.Status != 0 {
		return errors.New(res.Error)
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
	docStr := C.CString(string(docBytes))
	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(docStr))
	res := ConvertAndFreeCResult(C.LensUp(C.uintptr_t(w.handle), cCollectionID, docStr))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(res.Value), &out); err != nil {
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
	docStr := C.CString(string(docBytes))
	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(docStr))
	res := ConvertAndFreeCResult(C.LensDown(C.uintptr_t(w.handle), cCollectionID, docStr))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var out []map[string]any
	if err := json.Unmarshal([]byte(res.Value), &out); err != nil {
		return nil, err
	}
	return enumerable.New(out), nil
}
