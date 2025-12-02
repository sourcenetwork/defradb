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
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encoding"
	"github.com/sourcenetwork/defradb/internal/keys"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// storeArtifacts stores SE artifacts directly in the datastore.
func storeArtifacts(ctx context.Context, ds corekv.ReaderWriter, artifacts []secore.Artifact) error {
	for _, artifact := range artifacts {
		colID, err := id.GetShortCollectionID(ctx, artifact.CollectionID)
		if err != nil {
			return err
		}

		key := keys.DatastoreSE{
			CollectionShortID: colID,
			IndexID:           artifact.IndexID,
			SearchTag:         artifact.SearchTag,
			DocID:             artifact.DocID,
		}

		if err := ds.Set(ctx, key.Bytes(), []byte{}); err != nil {
			return err
		}
	}

	return nil
}

// fetchDocIDs queries the datastore for SE artifacts matching the given queries
// and returns the document IDs for documents that match all queries.
func fetchDocIDs(
	ctx context.Context,
	ds corekv.ReaderWriter,
	collectionID string,
	queries []fieldQuery,
) ([]string, error) {
	docIDSet := make(map[string]struct{})

	colID, err := id.GetShortCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	isFirstPass := true
	for _, query := range queries {
		key := keys.DatastoreSE{
			CollectionShortID: colID,
			IndexID:           query.IndexID,
			SearchTag:         query.SearchTag,
		}

		iter, err := ds.Iterator(ctx, corekv.IterOptions{
			Prefix: key.Bytes(),
		})
		if err != nil {
			return nil, err
		}

		querySet := make(map[string]struct{})
		for {
			hasNext, err := iter.Next()
			if err != nil || !hasNext {
				break
			}

			dsKey, err := keys.NewDatastoreSEFromString(string(iter.Key()))
			if err != nil {
				return nil, errors.Join(err, iter.Close())
			}
			if dsKey.DocID == "" {
				return nil, errors.Join(NewErrEmptyDocID(dsKey.ToString()), iter.Close())
			}

			querySet[dsKey.DocID] = struct{}{}
		}

		err = iter.Close()
		if err != nil {
			return nil, err
		}

		if isFirstPass {
			docIDSet = querySet
			isFirstPass = false
		} else {
			for docID := range docIDSet {
				if _, exists := querySet[docID]; !exists {
					delete(docIDSet, docID)
				}
			}
		}

		if len(docIDSet) == 0 {
			break
		}
	}

	docIDs := make([]string, 0, len(docIDSet))
	for docID := range docIDSet {
		docIDs = append(docIDs, docID)
	}

	return docIDs, nil
}

// fieldQuery represents a query for a specific encrypted field
type fieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// generateDocArtifacts generates SE artifacts for specified fields of a document.
// If fieldNames is empty or nil, artifacts are generated for all encrypted fields.
// The identity is used in tag computation to isolate tags per identity on the remote node.
func generateDocArtifacts(
	ctx context.Context,
	col client.Collection,
	doc *client.Document,
	fieldNames []string,
	identity immutable.Option[acpIdentity.Identity],
	encKey []byte,
) ([]secore.Artifact, error) {
	encryptedIndexes, err := col.ListEncryptedIndexes(ctx)
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
		if !slices.Contains(fieldNames, encIdx.FieldName) {
			continue
		}

		fieldValue, err := doc.GetValue(encIdx.FieldName)
		if err != nil {
			return nil, err
		}

		normalValue := fieldValue.NormalValue()
		artifact, err := generateFieldArtifact(collectionID, docID, encIdx, normalValue, identity, encKey)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// generateFieldArtifact generates a single SE artifact for a specific field value.
// The identity is included in the tag computation to isolate data per identity on remote nodes.
func generateFieldArtifact(
	collectionID string,
	docID string,
	encIdx client.EncryptedIndexDescription,
	fieldValue client.NormalValue,
	identity immutable.Option[acpIdentity.Identity],
	encKey []byte,
) (secore.Artifact, error) {
	valueBytes := encoding.EncodeFieldValue(nil, fieldValue, false)

	var identityStr string
	if identity.HasValue() {
		ident := identity.Value()
		if pubKey := ident.PublicKey(); pubKey != nil {
			identityStr = string(pubKey.Raw())
		}
	}

	var tag []byte
	switch encIdx.Type {
	case client.EncryptedIndexTypeEquality:
		tag = secore.GenerateEqualityTag(encKey, identityStr, collectionID, encIdx.FieldName, valueBytes)

	default:
		return secore.Artifact{}, NewErrUnsupportedIndexType(string(encIdx.Type))
	}

	artifact := secore.Artifact{
		CollectionID: collectionID,
		DocID:        docID,
		IndexID:      encIdx.FieldName,
		SearchTag:    tag,
	}

	return artifact, nil
}
