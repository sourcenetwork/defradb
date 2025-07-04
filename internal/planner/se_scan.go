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
	"context"
	"fmt"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/id"
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

	// Inner scan node for fetching documents
	innerScan *scanNode
}

var _ planNode = (*seScanNode)(nil)

func (n *seScanNode) Kind() string { return "seScanNode" }

func (n *seScanNode) Init() error {
	mapperSelect := &mapper.Select{
		CollectionName:  n.collection.Name(),
		DocumentMapping: n.documentMapping,
	}

	var err error
	n.innerScan, err = n.p.Scan(mapperSelect)
	if err != nil {
		return err
	}
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
	/*msg, responseChan := se.NewQuerySEArtifactsMessage(n.collectionID, queries)
	n.p.db.Events().Publish(msg)

	response := <-responseChan
	if response.Error != nil {
		return nil, response.Error
	}

	return response.DocIDs, nil*/
	return nil, nil
}

func (n *seScanNode) Next() (bool, error) {
	tries := 0
	for n.currentIndex < len(n.remoteDocIDs) && tries < 3 {
		fetched, err := n.fetchCurrentDocument()
		if err != nil {
			return false, err
		}
		if fetched {
			n.currentIndex++
			return true, nil
		}
		tries++
	}
	return false, nil
}

func (n *seScanNode) fetchCurrentDocument() (bool, error) {
	docIDStr := n.remoteDocIDs[n.currentIndex]

	doc, err := n.fetchDocumentLocallyByID(docIDStr)
	if err != nil {
		return false, err
	}

	if !doc.HasValue() {
		found, err := n.requestDocumentFromNetwork(docIDStr)
		if err != nil {
			return false, err
		}

		if !found {
			return false, nil
		}
	}

	n.currentValue = doc.Value()
	return true, nil
}

func (n *seScanNode) requestDocumentFromNetwork(docIDStr string) (bool, error) {
	responseChan := make(chan event.DocUpdateResponse, 1)
	defer close(responseChan)

	/*request := event.DocUpdateRequest{
		CollectionID: n.collectionID,
		DocID:        docIDStr,
		Response:     responseChan,
	}*/

	// TODO: waiting for every single document is not efficient.
	// We should consider ways of prefetching docs.
	//n.p.db.Events().Publish(event.NewMessage(event.DocUpdateRequestName, request))

	ctx, cancel := context.WithTimeout(n.p.ctx, 10*time.Second)
	defer cancel()

	select {
	case response := <-responseChan:
		if response.Error != nil {
			return false, response.Error
		}
		return response.Found, nil
	case <-ctx.Done():
		return false, fmt.Errorf("timeout waiting for document %s", docIDStr)
	}
}

func (n *seScanNode) fetchDocumentLocallyByID(docIDStr string) (immutable.Option[core.Doc], error) {
	shortID, err := id.GetShortCollectionID(n.p.ctx, n.collection.Version().CollectionID)
	if err != nil {
		return immutable.None[core.Doc](), err
	}

	dsKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             docIDStr,
	}

	prefixes := []keys.Walkable{dsKey}
	n.innerScan.Prefixes(prefixes)

	if err := n.innerScan.Init(); err != nil {
		return immutable.None[core.Doc](), err
	}

	hasValue, err := n.innerScan.Next()
	if err != nil || !hasValue {
		return immutable.None[core.Doc](), err
	}

	return immutable.Some(n.innerScan.Value()), nil
}

func (n *seScanNode) Prefixes(prefixes []keys.Walkable) {}

func (n *seScanNode) Source() planNode { return nil }

func (n *seScanNode) Close() error {
	if n.innerScan != nil {
		return n.innerScan.Close()
	}
	return nil
}
