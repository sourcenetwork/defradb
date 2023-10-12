// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package descriptions

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/cid"
	"github.com/sourcenetwork/defradb/datastore"
)

// CreateSchemaVersion creates and saves to the store a new schema version.
//
// If the SchemaID is empty it will be set to the new version ID.
func CreateSchemaVersion(
	ctx context.Context,
	txn datastore.Txn,
	desc client.SchemaDescription,
) (client.SchemaDescription, error) {
	for i := range desc.Fields {
		// This is not wonderful and will probably break when we add the ability
		// to delete fields, however it is good enough for now and matches the
		// create behaviour.
		desc.Fields[i].ID = client.FieldID(i)
	}

	buf, err := json.Marshal(desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	scid, err := cid.NewSHA256CidV1(buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	versionID := scid.String()
	previousSchemaVersionID := desc.VersionID
	isNew := desc.SchemaID == ""

	desc.VersionID = versionID
	if isNew {
		// If this is a new schema, the schema ID will match the version ID
		desc.SchemaID = versionID
	}

	// Rebuild the json buffer to include the newly set ID properties
	buf, err = json.Marshal(desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	key := core.NewSchemaVersionKey(versionID)
	err = txn.Systemstore().Put(ctx, key.ToDS(), buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	if !isNew {
		// We don't need to add a history key if this is the first version
		schemaVersionHistoryKey := core.NewSchemaHistoryKey(desc.SchemaID, previousSchemaVersionID)
		err = txn.Systemstore().Put(ctx, schemaVersionHistoryKey.ToDS(), []byte(desc.VersionID))
		if err != nil {
			return client.SchemaDescription{}, err
		}
	}

	return desc, nil
}
