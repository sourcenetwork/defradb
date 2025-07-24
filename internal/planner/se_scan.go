// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/internal/se"
)

// seScanNode implements a plan node for searchable encryption queries.
// It queries remote nodes for document IDs matching the search criteria
// and returns only the document IDs.
type seScanNode struct {
	documentIterator
	docMapper

	p                *Planner
	collection       client.Collection
	collectionID     string
	filter           *mapper.Filter
	encryptedIndexes []client.EncryptedIndexDescription

	fieldSearchTags map[string][]byte
	remoteDocIDs    []string
	hasReturned     bool
}

var _ planNode = (*seScanNode)(nil)

func (n *seScanNode) Kind() string { return "seScanNode" }

func (n *seScanNode) Init() error {
	return nil
}

func (n *seScanNode) Start() error {
	if err := n.generateSearchTags(); err != nil {
		return err
	}

	n.remoteDocIDs = nil
	n.hasReturned = false

	return nil
}

func (n *seScanNode) generateSearchTags() error {
	n.fieldSearchTags = make(map[string][]byte)

	for fieldName, condition := range n.filter.ExternalConditions {

		var encIdx *client.EncryptedIndexDescription
		for _, idx := range n.encryptedIndexes {
			if idx.FieldName == fieldName {
				encIdx = &idx
				break
			}
		}

		if encIdx == nil {
			continue
		}

		condMap, ok := condition.(map[string]any)
		if !ok {
			return NewErrInvalidEncryptedFieldCondition(fieldName)
		}

		value, hasEq := condMap["_eq"]
		if !hasEq {
			return NewErrUnsupportedEncryptedOperator(fieldName)
		}

		normalValue, err := client.NewNormalValue(value)
		if err != nil {
			return NewErrFailedToCreateNormalValue(fieldName, err)
		}

		artifact, err := se.GenerateFieldArtifact(
			n.p.ctx,
			n.collectionID,
			"",
			*encIdx,
			normalValue,
			n.p.db.GetSearchableEncryptionKey(),
		)
		if err != nil {
			return NewErrFailedToGenerateSearchTag(fieldName, err)
		}

		n.fieldSearchTags[fieldName] = artifact.SearchTag
	}

	return nil
}

func (n *seScanNode) queryRemoteNodes() ([]string, error) {
	if len(n.fieldSearchTags) == 0 {
		return nil, nil
	}

	queries := make([]se.FieldQuery, 0, len(n.fieldSearchTags))
	for fieldName, searchTag := range n.fieldSearchTags {
		queries = append(queries, se.FieldQuery{
			FieldName: fieldName,
			IndexID:   fieldName,
			SearchTag: searchTag,
		})
	}

	msg, responseChan := se.NewQuerySEArtifactsMessage(n.collectionID, queries)
	n.p.db.Events().Publish(msg)

	response := <-responseChan
	if response.Error != nil {
		return nil, response.Error
	}

	return response.DocIDs, nil
}

func (n *seScanNode) Next() (bool, error) {
	if n.hasReturned {
		return false, nil
	}

	if n.remoteDocIDs == nil {
		docIDs, err := n.queryRemoteNodes()
		if err != nil {
			return false, err
		}
		n.remoteDocIDs = docIDs
	}

	doc := n.documentMapping.NewDoc()
	n.documentMapping.SetFirstOfName(&doc, request.DocIDsFieldName, n.remoteDocIDs)
	n.currentValue = doc
	n.hasReturned = true

	return true, nil
}

func (n *seScanNode) Prefixes(prefixes []keys.Walkable) {}

func (n *seScanNode) Source() planNode { return nil }

func (n *seScanNode) Close() error {
	return nil
}
