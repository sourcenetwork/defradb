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
#include <stdint.h>
#include "defra_structs.h"
extern Result PurgeDocuments(uintptr_t nodePtr, char* collectionName, char* docIDsJSON,
int pruneHistory, uintptr_t identityPtr);
extern void FreeIdentity(uintptr_t identityPtr);
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (w *CWrapper) PurgeDocuments(
	ctx context.Context,
	collectionName client.CollectionName,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.PurgeDocumentsOptions],
) error {
	rawIDs := make([]string, len(docIDs))
	for i, docID := range docIDs {
		rawIDs[i] = docID.String()
	}
	encodedIDs, err := json.Marshal(rawIDs)
	if err != nil {
		return err
	}

	cName := C.CString(collectionName)
	cDocIDs := C.CString(string(encodedIDs))
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cPruneHistory := C.int(0)
	if pruneHistory {
		cPruneHistory = 1
	}

	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cDocIDs))
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(
		C.PurgeDocuments(
			callHandle,
			cName,
			cDocIDs,
			cPruneHistory,
			cIdentity,
		),
	)
	if res.Status != 0 {
		return errors.New(res.Error)
	}

	return nil
}
