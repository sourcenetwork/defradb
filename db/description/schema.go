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

	"github.com/ipfs/go-datastore/query"

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
	buf, err := json.Marshal(desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	scid, err := cid.NewSHA256CidV1(buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	versionID := scid.String()
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
	err = txn.Systemstore().Put(ctx, key.ToDS(), buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	if !isNew {
		// We don't need to add a root key if this is the first version
		schemaVersionHistoryKey := core.NewSchemaRootKey(desc.Root, desc.VersionID)
		err = txn.Systemstore().Put(ctx, schemaVersionHistoryKey.ToDS(), []byte{})
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

	buf, err := txn.Systemstore().Get(ctx, key.ToDS())
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
	cols, err := GetActiveCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	versionIDs := make([]string, 0)
	for _, col := range cols {
		versionIDs = append(versionIDs, col.SchemaVersionID)
	}

	schemaVersionPrefix := core.NewSchemaVersionKey("")
	schemaVersionQuery, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: schemaVersionPrefix.ToString(),
	})
	if err != nil {
		return nil, NewErrFailedToCreateSchemaQuery(err)
	}

	descriptions := make([]client.SchemaDescription, 0)
	for res := range schemaVersionQuery.Next() {
		if res.Error != nil {
			if err := schemaVersionQuery.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		var desc client.SchemaDescription
		err = json.Unmarshal(res.Value, &desc)
		if err != nil {
			if err := schemaVersionQuery.Close(); err != nil {
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

	if err := schemaVersionQuery.Close(); err != nil {
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
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, NewErrFailedToCreateSchemaQuery(err)
	}

	schemas := make([]client.SchemaDescription, 0)
	for res := range q.Next() {
		if res.Error != nil {
			if err := q.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		var desc client.SchemaDescription
		err = json.Unmarshal(res.Value, &desc)
		if err != nil {
			if err := q.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		schemas = append(schemas, desc)
	}

	if err := q.Close(); err != nil {
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

	prefix := core.NewSchemaRootKey(schemaRoot, "")
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrFailedToCreateSchemaQuery(err)
	}

	for res := range q.Next() {
		if res.Error != nil {
			if err := q.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		key, err := core.NewSchemaRootKeyFromString(res.Key)
		if err != nil {
			return nil, err
		}

		schemaVersions = append(schemaVersions, key.SchemaVersionID)
	}

	return schemaVersions, nil
}

// GetCollectionlessSchemas returns all schema that are not attached to a collection.
//
// Typically this means any schema embedded in a View.
//
// WARNING: This function does not currently account for multiple versions of collectionless schema,
// at the moment such a situation is impossible, but that is likely to change, at which point this
// function will need to account for that.
func GetCollectionlessSchemas(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.SchemaDescription, error) {
	cols, err := GetCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	allSchemas, err := GetAllSchemas(ctx, txn)
	if err != nil {
		return nil, err
	}

	schemaRootsByVersionID := map[string]string{}
	for _, schema := range allSchemas {
		schemaRootsByVersionID[schema.VersionID] = schema.Root
	}

	colSchemaRoots := map[string]struct{}{}
	for _, col := range cols {
		schemaRoot := schemaRootsByVersionID[col.SchemaVersionID]
		colSchemaRoots[schemaRoot] = struct{}{}
	}

	collectionlessSchema := []client.SchemaDescription{}
	for _, schema := range allSchemas {
		if _, hasCollection := colSchemaRoots[schema.Root]; !hasCollection {
			collectionlessSchema = append(collectionlessSchema, schema)
		}
	}

	return collectionlessSchema, nil
}
