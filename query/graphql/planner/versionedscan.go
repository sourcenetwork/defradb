package planner

import (
	"errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
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
	key     core.Key
	version cid.Cid

	desc  base.CollectionDescription
	index *base.IndexDescription

	fields []*base.FieldDescription
	doc    map[string]interface{}
	docKey []byte

	// map between fieldID and index in fields
	fieldIdxMap map[base.FieldID]int

	spans            core.Spans
	isSecondaryIndex bool
	reverse          bool

	// rowIndex int64

	// filter data
	filter *parser.Filter

	scanInitialized bool

	fetcher fetcher.VersionedFetcher
}

func (n *versionedScanNode) Init() error {
	// init the fetcher
	if err := n.fetcher.Init(&n.desc, nil, n.fields, n.reverse); err != nil {
		return err
	}
	return n.initScan()
}

func (n *versionedScanNode) initCollection(desc base.CollectionDescription) error {
	n.desc = desc
	n.index = &desc.Indexes[0]
	return nil
}

func (n *versionedScanNode) initVersion(key core.Key, c cid.Cid) error {
	n.key = key
	n.version = c
	return nil
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *versionedScanNode) Start() error {
	return nil // noop
}

func (n *versionedScanNode) initScan() error {
	if len(n.key.String()) == 0 || n.version.Equals(emptyCID) {
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
