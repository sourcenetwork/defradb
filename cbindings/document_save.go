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
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/encryption"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export SaveDocument
func SaveDocument(
	nodePtr C.uintptr_t,
	docIDStr *C.char,
	jsonData *C.char,
	isEncrypted C.int,
	encryptedFields *C.char,
	opts C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colOptions := parseCollectionOptionsToGetCollectionsOptions(opts)
	ident := acpIdentity.FromContext(ctx)
	if ident.HasValue() {
		colOptions.SetIdentity(ident.Value())
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var encryptFields []string
	encryptFieldsStr := C.GoString(encryptedFields)
	if encryptFieldsStr != "" {
		for _, f := range strings.Split(encryptFieldsStr, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				encryptFields = append(encryptFields, trimmed)
			}
		}
	}
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted != 0, encryptFields)

	saveOpt := options.WithIdentity(options.SaveDocument(), acpIdentity.FromContext(ctx))
	if opts.enableSigning != 0 {
		saveOpt.SetEnableSigning(opts.enableSigning > 0)
	}

	jsonString := strings.TrimSpace(C.GoString(jsonData))
	docID := C.GoString(docIDStr)
	var doc *client.Document
	if docID != "" {
		newDocID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		doc, err = client.NewDocWithID(ctx, newDocID, col.Version())
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		if len(jsonString) > 0 {
			if err := doc.SetWithJSON(ctx, []byte(jsonString)); err != nil {
				return returnC(returnGoC(1, err.Error(), ""))
			}
		}
	} else {
		var rawMap map[string]any
		if err := json.Unmarshal([]byte(jsonString), &rawMap); err == nil && rawMap != nil {
			if _, hasDocID := rawMap[request.DocIDFieldName]; hasDocID {
				doc, err = client.NewDocFromMap(ctx, rawMap, col.Version())
				if err != nil {
					return returnC(returnGoC(1, err.Error(), ""))
				}
			}
		}
		if doc == nil {
			doc, err = client.NewDocFromJSON(ctx, []byte(jsonString), col.Version())
			if err != nil {
				return returnC(returnGoC(1, err.Error(), ""))
			}
		}
	}

	err = col.SaveDocument(ctx, doc, saveOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	docIDs, err := json.Marshal(client.DocumentIDs([]*client.Document{doc}))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", string(docIDs)))
}
