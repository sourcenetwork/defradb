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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// CreateSchemaVersion creates and saves to the store a new schema version.
//
// If the Root is empty it will be set to the new version ID.
func CreateSchemaVersion(
	ctx context.Context,
	desc client.SchemaDescription,
) (client.SchemaDescription, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	buf, err := json.Marshal(desc)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	key := keys.NewSchemaVersionKey(desc.VersionID)
	err = txn.Systemstore().Set(ctx, key.Bytes(), buf)
	if err != nil {
		return client.SchemaDescription{}, err
	}

	isNew := desc.Root == desc.VersionID
	if !isNew {
		// We don't need to add a root key if this is the first version
		schemaVersionHistoryKey := keys.NewSchemaRootKey(desc.Root, desc.VersionID)
		err = txn.Systemstore().Set(ctx, schemaVersionHistoryKey.Bytes(), []byte{})
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
	versionID string,
) (client.SchemaDescription, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	key := keys.NewSchemaVersionKey(versionID)

	buf, err := txn.Systemstore().Get(ctx, key.Bytes())
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
	name string,
) ([]client.SchemaDescription, error) {
	allSchemas, err := GetAllSchemas(ctx)
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
	root string,
) ([]client.SchemaDescription, error) {
	allSchemas, err := GetAllSchemas(ctx)
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
) ([]client.SchemaDescription, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	cols, err := GetActiveCollections(ctx)
	if err != nil {
		return nil, err
	}

	versionIDs := make([]string, 0)
	for _, col := range cols {
		versionIDs = append(versionIDs, col.VersionID)
	}

	schemaVersionPrefix := keys.NewSchemaVersionKey("")
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: schemaVersionPrefix.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	descriptions := make([]client.SchemaDescription, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		var desc client.SchemaDescription
		err = json.Unmarshal(value, &desc)
		if err != nil {
			if err := iter.Close(); err != nil {
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

	if err := iter.Close(); err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	return descriptions, nil
}

// GetSchemas returns all schema versions in the system.
func GetAllSchemas(
	ctx context.Context,
) ([]client.SchemaDescription, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewSchemaVersionKey("").Bytes(),
	})
	if err != nil {
		return nil, err
	}

	schemas := make([]client.SchemaDescription, 0)
	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		value, err := iter.Value()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		var desc client.SchemaDescription
		err = json.Unmarshal(value, &desc)
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		schemas = append(schemas, desc)
	}

	if err := iter.Close(); err != nil {
		return nil, NewErrFailedToCloseSchemaQuery(err)
	}

	return schemas, nil
}

func GetSchemaVersionIDs(
	ctx context.Context,
	schemaRoot string,
) ([]string, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	// Add the schema root as the first version here.
	// It is not present in the history prefix.
	schemaVersions := []string{schemaRoot}

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewSchemaRootKey(schemaRoot, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	for {
		hasValue, err := iter.Next()
		if err != nil {
			if err := iter.Close(); err != nil {
				return nil, NewErrFailedToCloseSchemaQuery(err)
			}
			return nil, err
		}

		if !hasValue {
			break
		}

		key, err := keys.NewSchemaRootKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		schemaVersions = append(schemaVersions, key.SchemaVersionID)
	}

	return schemaVersions, iter.Close()
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
) ([]client.SchemaDescription, error) {
	cols, err := GetCollections(ctx)
	if err != nil {
		return nil, err
	}

	allSchemas, err := GetAllSchemas(ctx)
	if err != nil {
		return nil, err
	}

	schemaRootsByVersionID := map[string]string{}
	for _, schema := range allSchemas {
		schemaRootsByVersionID[schema.VersionID] = schema.Root
	}

	colSchemaRoots := map[string]struct{}{}
	for _, col := range cols {
		schemaRoot := schemaRootsByVersionID[col.VersionID]
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
