// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SetFieldIDs sets the field IDs hosted on the given collections, mutating the input set.
func SetFieldIDs(ctx context.Context, txn datastore.Txn, definitions []client.CollectionDefinition) error {
	for i := range definitions {
		fieldSeq, err := sequence.Get(ctx, txn, keys.NewFieldIDSequenceKey(definitions[i].Version.CollectionID))
		if err != nil {
			return err
		}

		for j := range definitions[i].Version.Fields {
			var fieldID client.FieldID
			if definitions[i].Version.Fields[j].ID != client.FieldID(0) {
				fieldID = definitions[i].Version.Fields[j].ID
			} else if definitions[i].Version.Fields[j].Name == request.DocIDFieldName {
				// There is no hard technical requirement for this, we just think it looks nicer
				// if the doc id is at the zero index.  It makes it look a little nicer in commit
				// queries too.
				fieldID = 0
			} else {
				nextID, err := fieldSeq.Next(ctx, txn)
				if err != nil {
					return err
				}
				fieldID = client.FieldID(nextID)
			}

			definitions[i].Version.Fields[j].ID = fieldID
		}
	}

	return nil
}
