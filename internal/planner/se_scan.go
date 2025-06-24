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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/internal/se"
)

// seScanNode implements a plan node for searchable encryption queries.
// It queries remote nodes for document IDs matching the search criteria
// and then fetches the full documents from the regular P2P network.
type seScanNode struct {
	documentIterator
	docMapper

	p                *Planner
	collection       client.Collection
	collectionID     string
	filter           *mapper.Filter
	encryptedIndexes []client.EncryptedIndexDescription

	// SE specific fields
	fieldSearchTags map[string][]byte
	remoteDocIDs    []string
	currentIndex    int
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

	docIDs, err := n.queryRemoteNodes()
	if err != nil {
		return err
	}

	n.remoteDocIDs = docIDs
	n.currentIndex = 0

	return nil
}

func (n *seScanNode) generateSearchTags() error {
	n.fieldSearchTags = make(map[string][]byte)

	for filterKey, condition := range n.filter.Conditions {
		objProp, ok := filterKey.(*mapper.ObjectProperty)
		if !ok {
			continue
		}

		fieldName := objProp.Name

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
			return fmt.Errorf("invalid condition for encrypted field %s", fieldName)
		}

		value, hasEq := condMap["_eq"]
		if !hasEq {
			return fmt.Errorf("only _eq operator supported for encrypted field %s", fieldName)
		}

		normalValue, err := client.NewNormalValue(value)
		if err != nil {
			return fmt.Errorf("failed to create normal value for field %s: %w", fieldName, err)
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
			return fmt.Errorf("failed to generate search tag for field %s: %w", fieldName, err)
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

	// Create and publish the query request
	// This event will be handled by the network layer which will:
	// 1. Query replicator nodes for matching SE artifacts
	// 2. Aggregate results from multiple nodes
	// 3. Send the response back via the channel
	msg, responseChan := se.NewQuerySEArtifactsMessage(n.collectionID, queries)
	n.p.db.Events().Publish(msg)

	response := <-responseChan
	if response.Error != nil {
		return nil, response.Error
	}

	return response.DocIDs, nil
}

func (n *seScanNode) Next() (bool, error) {
	if n.currentIndex >= len(n.remoteDocIDs) {
		return false, nil
	}

	// Get the next document ID
	docIDStr := n.remoteDocIDs[n.currentIndex]
	docID, err := client.NewDocIDFromString(docIDStr)
	if err != nil {
		n.currentIndex++
		return n.Next() // Skip invalid doc ID
	}

	// First, try to get the document from local store (in case it exists)
	doc, err := n.collection.Get(n.p.ctx, docID, false)
	if err == nil {
		// Document exists locally, use it
		return n.processDocument(doc)
	}

	// Document not found locally, need to request it from the network
	// TODO: Implement document request from P2P network
	// This will require:
	// 1. Publishing a document request on pubsub
	// 2. Waiting for pushLog response
	// 3. Processing the pushLog to trigger merge
	// 4. Waiting for merge completion

	// For now, skip documents that aren't locally available
	n.currentIndex++
	return n.Next()
}

func (n *seScanNode) processDocument(doc *client.Document) (bool, error) {
	// Convert document to core.Doc format
	fields := make(core.DocFields, 0)

	// Add DocID as first field
	fields = append(fields, doc.ID().String())

	// TODO: Add proper field mapping based on document mapping
	// For now, create a simple Doc structure
	n.currentValue = core.Doc{
		Fields: fields,
		Status: client.Active,
	}
	n.currentIndex++

	return true, nil
}

func (n *seScanNode) Prefixes(prefixes []keys.Walkable) {}

func (n *seScanNode) Source() planNode { return nil }

func (n *seScanNode) Close() error { return nil }
