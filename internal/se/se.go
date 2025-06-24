// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package se provides searchable encryption support.
*/
package se

import (
	"context"
	"slices"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/encoding"
	"github.com/sourcenetwork/defradb/internal/keys"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// StoreArtifacts stores SE artifacts directly in the datastore.
// This is called by the server when receiving artifacts from peers.
func StoreArtifacts(ctx context.Context, ds datastore.DSReaderWriter, artifacts []secore.Artifact) error {
	for _, artifact := range artifacts {
		key := keys.DatastoreSE{
			CollectionID: artifact.CollectionID,
			IndexID:      artifact.IndexID,
			SearchTag:    artifact.SearchTag,
			DocID:        artifact.DocID,
		}

		// Store empty value - we only need the key for search lookups
		if err := ds.Set(ctx, key.Bytes(), []byte{}); err != nil {
			return err
		}
	}

	return nil
}

// FetchDocIDs queries the datastore for SE artifacts matching the given queries
// and returns the unique document IDs that match.
func FetchDocIDs(ctx context.Context, ds datastore.DSReaderWriter, collectionID string, queries []FieldQuery) ([]string, error) {
	docIDSet := make(map[string]struct{})

	for _, query := range queries {
		key := keys.DatastoreSE{
			CollectionID: collectionID,
			IndexID:      query.IndexID,
			SearchTag:    query.SearchTag,
		}

		iter, err := ds.Iterator(ctx, corekv.IterOptions{
			Prefix: key.Bytes(),
		})
		if err != nil {
			return nil, err
		}
		defer iter.Close()

		for {
			hasNext, err := iter.Next()
			if err != nil || !hasNext {
				break
			}

			dsKey, err := keys.NewDatastoreSEFromString(string(iter.Key()))
			if err != nil {
				return nil, err
			}
			if dsKey.DocID == "" {
				return nil, NewErrEmptyDocID(dsKey.ToString())
			}

			docIDSet[dsKey.DocID] = struct{}{}
		}
	}

	docIDs := make([]string, 0, len(docIDSet))
	for docID := range docIDSet {
		docIDs = append(docIDs, docID)
	}

	return docIDs, nil
}

// FieldQuery represents a query for a specific encrypted field
type FieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// GenerateDocArtifacts generates SE artifacts for specified fields of a document.
// If fieldNames is empty or nil, artifacts are generated for all encrypted fields.
func GenerateDocArtifacts(
	ctx context.Context,
	col client.Collection,
	doc *client.Document,
	fieldNames []string,
	encKey []byte,
) ([]secore.Artifact, error) {
	encryptedIndexes, err := col.GetEncryptedIndexes(ctx)
	if err != nil {
		return nil, NewErrFailedToGetEncryptedIndexes(err)
	}

	if len(encryptedIndexes) == 0 {
		return nil, nil
	}

	collectionID := col.VersionID()
	docID := doc.ID().String()

	var artifacts []secore.Artifact
	for _, encIdx := range encryptedIndexes {
		// Skip if fieldNames is specified and this field is not in the list
		if len(fieldNames) > 0 && !slices.Contains(fieldNames, encIdx.FieldName) {
			continue
		}

		fieldValue, err := doc.GetValue(encIdx.FieldName)
		if err != nil {
			return nil, NewErrFailedToGetFieldValue(encIdx.FieldName, err)
		}

		normalValue := fieldValue.NormalValue()
		artifact, err := GenerateFieldArtifact(ctx, collectionID, docID, encIdx, normalValue, encKey)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// GenerateFieldArtifact generates a single SE artifact for a specific field value.
// This function encapsulates the logic for creating search tags and artifacts
// for individual fields, making it reusable across different contexts.
func GenerateFieldArtifact(
	ctx context.Context,
	collectionID string,
	docID string,
	encIdx client.EncryptedIndexDescription,
	fieldValue client.NormalValue,
	encKey []byte,
) (secore.Artifact, error) {
	valueBytes := encoding.EncodeFieldValue(nil, fieldValue, false)

	var tag []byte
	switch encIdx.Type {
	case client.EncryptedIndexTypeEquality:
		var err error
		tag, err = secore.GenerateEqualityTag(
			encKey,
			collectionID,
			encIdx.FieldName,
			valueBytes,
		)

		if err != nil {
			log.ErrorContextE(ctx, "Failed to generate search tag", err,
				corelog.String("FieldName", encIdx.FieldName))
			return secore.Artifact{}, err
		}

	default:
		log.ErrorContext(ctx, "Unsupported encrypted index type",
			corelog.String("Type", string(encIdx.Type)))
		return secore.Artifact{}, NewErrUnsupportedIndexType(string(encIdx.Type))
	}

	artifact := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: collectionID,
		FieldName:    encIdx.FieldName,
		DocID:        docID,
		Operation:    secore.OperationAdd,
		IndexID:      encIdx.FieldName,
		SearchTag:    tag,
	}

	return artifact, nil
}
