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

	remoteDocIDs []string
	hasReturned  bool
}

var _ planNode = (*seScanNode)(nil)

func (n *seScanNode) Kind() string { return "seScanNode" }

func (n *seScanNode) Init() error {
	return nil
}

func (n *seScanNode) Start() error {
	n.remoteDocIDs = nil
	n.hasReturned = false
	return nil
}

func (n *seScanNode) queryRemoteNodes() ([]string, error) {
	fieldValues := make([]se.FieldValueQuery, 0, len(n.filter.ExternalConditions))

	for fieldName, condition := range n.filter.ExternalConditions {
		// Find the encrypted index for this field
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

		// Extract the equality value
		value, hasEq := condition.(map[string]any)["_eq"]
		if !hasEq {
			return nil, NewErrUnsupportedEncryptedOperator(fieldName)
		}

		// Create normal value
		normalValue, err := client.NewNormalValue(value)
		if err != nil {
			return nil, NewErrFailedToCreateNormalValue(fieldName, err)
		}

		fieldValues = append(fieldValues, se.FieldValueQuery{
			FieldName: fieldName,
			IndexDesc: *encIdx,
			Value:     normalValue,
		})
	}

	docIDs, err := n.p.p2p.QueryDocIDsWithSETags(
		n.p.ctx,
		n.collectionID,
		fieldValues,
	)
	if err != nil {
		return nil, err
	}

	return docIDs, nil
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
