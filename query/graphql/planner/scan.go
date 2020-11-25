package planner

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// scans an index for records
type scanNode struct {
	p     *Planner
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

	fetcher fetcher.DocumentFetcher
}

func (n *scanNode) Init() error {
	return n.initScan()
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	// init the fetcher
	return n.fetcher.Init(&n.desc, n.index, n.fields, n.reverse)
}

func (n *scanNode) initScan() error {
	if len(n.spans) == 0 {
		start := base.MakeIndexPrefixKey(&n.desc, n.index)
		n.spans = append(n.spans, core.NewSpan(start, start.PrefixEnd()))
	}

	err := n.fetcher.Start(n.p.txn, n.spans)
	if err != nil {
		return err
	}

	n.scanInitialized = true
	return nil
}

// Next gets the next result.
// Returns true, if there is a result,
// and false otherwise.
func (n *scanNode) Next() (bool, error) {
	if !n.scanInitialized {
		if err := n.initScan(); err != nil {
			return false, err
		}
	}

	// keep scanning until we find a doc that passes the filter
	for {
		var err error
		n.docKey, n.doc, err = n.fetcher.FetchNextMap()
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

func (n *scanNode) Spans(spans core.Spans) {
	n.spans = spans
}

// Values returns the most recent result from Next()
func (n *scanNode) Values() map[string]interface{} {
	return n.doc
}

func (n *scanNode) Close() {}
