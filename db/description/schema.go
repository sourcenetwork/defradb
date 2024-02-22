// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/cid"
	"github.com/sourcenetwork/defradb/datastore"
)

// CreateSchemaVersion creates and saves to the store a new schema version.
//
// If the Root is empty it will be set to the new version ID.
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
	isNew := desc.Root == ""

	desc.VersionID = versionID
	if isNew {
		// If this is a new schema, the Root will match the version ID
		desc.Root = versionID
	}

	// Rebuild the json buffer to include the newly set ID properties
	buf, err = json.Marshal(desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	key := core.NewSchemaVersionKey(versionID)
	err = txn.Systemstore().Set(ctx, key.ToDS().Bytes(), buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	if !isNew {
		// We don't need to add a history key if this is the first version
		schemaVersionHistoryKey := core.NewSchemaHistoryKey(desc.Root, previousSchemaVersionID)
		err = txn.Systemstore().Set(ctx, schemaVersionHistoryKey.ToDS().Bytes(), []byte(desc.VersionID))
		if err != nil {
			return client.SchemaDescription{}, err
		}
	}

	return desc, nil
}

// GetSchemaVersion returns the schema description for the schema version of the
// ID provided.
//
// Will return an error if it is not found.
func GetSchemaVersion(
	ctx context.Context,
	txn datastore.Txn,
	versionID string,
) (client.SchemaDescription, error) {
	key := core.NewSchemaVersionKey(versionID)

	buf, err := txn.Systemstore().Get(ctx, key.ToDS().Bytes())
	if err != nil {
		return client.SchemaDescription{}, err
	}

	var desc client.SchemaDescription
	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	return desc, nil
}

// GetSchemasByName returns all the schema with the given name.
func GetSchemasByName(
	ctx context.Context,
	txn datastore.Txn,
	name string,
) ([]client.SchemaDescription, error) {
	allSchemas, err := GetAllSchemas(ctx, txn)
	if err != nil {
		return nil, err
	}

	nameSchemas := []client.SchemaDescription{}
	for _, schema := range allSchemas {
		if schema.Name == name {
			nameSchemas = append(nameSchemas, schema)
		}
	}

	return nameSchemas, nil
}

// GetSchemasByRoot returns all the schema with the given root.
func GetSchemasByRoot(
	ctx context.Context,
	txn datastore.Txn,
	root string,
) ([]client.SchemaDescription, error) {
	allSchemas, err := GetAllSchemas(ctx, txn)
	if err != nil {
		return nil, err
	}

	rootSchemas := []client.SchemaDescription{}
	for _, schema := range allSchemas {
		if schema.Root == root {
			rootSchemas = append(rootSchemas, schema)
		}
	}

	return rootSchemas, nil
}

// GetSchemas returns the schema of all the default schema versions in the system.
func GetSchemas(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.SchemaDescription, error) {
	cols, err := GetCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	versionIDs := make([]string, 0)
	for _, col := range cols {
		versionIDs = append(versionIDs, col.SchemaVersionID)
	}

	schemaVersionPrefix := core.NewSchemaVersionKey("")
	versionIter := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: schemaVersionPrefix.Bytes(),
	})

	descriptions := make([]client.SchemaDescription, 0)
	for ; versionIter.Valid(); versionIter.Next() {
		var desc client.SchemaDescription
		err = json.Unmarshal(versionIter.Value(), &desc)
		if err != nil {
			if err := versionIter.Close(ctx); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		for _, versionID := range versionIDs {
			if desc.VersionID == versionID {
				descriptions = append(descriptions, desc)
				break
			}
		}
	}

	if err := versionIter.Close(ctx); err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	return descriptions, nil
}

// GetSchemas returns all schema versions in the system.
func GetAllSchemas(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.SchemaDescription, error) {
	prefix := core.NewSchemaVersionKey("")
	iter := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: prefix.Bytes(),
	})

	schemas := make([]client.SchemaDescription, 0)
	for ; iter.Valid(); iter.Next() {
		var desc client.SchemaDescription
		err := json.Unmarshal(iter.Value(), &desc)
		if err != nil {
			if err := iter.Close(ctx); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		schemas = append(schemas, desc)
	}

	if err := iter.Close(ctx); err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	return schemas, nil
}

func GetSchemaVersionIDs(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
) ([]string, error) {
	// Add the schema root as the first version here.
	// It is not present in the history prefix.
	schemaVersions := []string{schemaRoot}

	prefix := core.NewSchemaHistoryKey(schemaRoot, "")
	iter := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix.Bytes(),
		KeysOnly: true,
	})

	for ; iter.Valid(); iter.Next() {
		key, err := core.NewSchemaHistoryKeyFromString(string(iter.Key()))
		if err != nil {
			if err := iter.Close(ctx); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		schemaVersions = append(schemaVersions, key.PreviousSchemaVersionID)
	}
	err := iter.Close(ctx)
	if err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	return schemaVersions, nil
}
