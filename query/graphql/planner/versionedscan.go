// Copyright 2022 Democratized Data Foundation
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
	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	"github.com/ipfs/go-cid"
)

var (
	emptyCID = cid.Cid{}
)

// scans an index for records
type versionedScanNode struct {
	p *Planner

	// versioned data
	key     core.DataStoreKey
	version cid.Cid

	desc client.CollectionDescription

	doc    map[string]interface{}
	docKey []byte

	// filter data
	filter *parser.Filter

	scanInitialized bool

	fetcher fetcher.VersionedFetcher
}

func (n *versionedScanNode) Init() error {
	// init the fetcher
	if err := n.fetcher.Init(&n.desc, nil); err != nil {
		return err
	}
	return n.initScan()
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *versionedScanNode) Start() error {
	return nil // noop
}

func (n *versionedScanNode) initScan() error {
	if n.key.DocKey == "" || n.version.Equals(emptyCID) {
		return errors.New("VersionedScan is missing either a DocKey or VersionCID")
	}

	// create a span of the form {DocKey, VersionCID}
	// spans := core.Spans{core.NewSpan(n.key, core.NewKey(n.version.String()))}
	spans := fetcher.NewVersionedSpan(n.key, n.version)
	err := n.fetcher.Start(n.p.ctx, n.p.txn, spans)
	if err != nil {
		return err
	}

	n.scanInitialized = true
	return nil
}

// Next gets the next result.
// Returns true, if there is a result,
// and false otherwise.
func (n *versionedScanNode) Next() (bool, error) {
	// if !n.scanInitialized {
	// 	if err := n.initScan(); err != nil {
	// 		return false, err
	// 	}
	// }

	// keep scanning until we find a doc that passes the filter
	for {
		var err error
		n.docKey, n.doc, err = n.fetcher.FetchNextMap(n.p.ctx)
		if err != nil {
			return false, err
		}
		if n.doc == nil {
			return false, nil
		}

		passed, err := parser.RunFilter(n.doc, n.filter, n.p.evalCtx)
		if err != nil {
			return false, err
		}
		if passed {
			return true, nil
		}
	}
}

func (n *versionedScanNode) Spans(spans core.Spans) {
	// n.spans = spans
	// we expect 1 span that includes both the DocKey and VersionCID
	// if len(spans) != 1 {
	// 	return
	// }
}

// Values returns the most recent result from Next()
func (n *versionedScanNode) Values() map[string]interface{} {
	return n.doc
}

func (n *versionedScanNode) Close() error {
	return n.fetcher.Close()
}

func (n *versionedScanNode) Source() planNode { return nil }

// Merge implements mergeNode
func (n *versionedScanNode) Merge() bool { return true }

func (p *Planner) VersionedScan() *versionedScanNode {
	return &versionedScanNode{p: p}
}
