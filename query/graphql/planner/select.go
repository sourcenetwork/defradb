package planner

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// wraps a selectNode and all the logic of a plan
// graph into a single struct for proper plan
// expansion
// Executes the top level plan node.
type selectTopNode struct {
	source planNode
	group  *groupNode
	sort   *sortNode
	limit  *limitNode
	render *renderNode

	// top of the plan graph
	plan planNode
}

func (n *selectTopNode) Init() error                    { return n.plan.Init() }
func (n *selectTopNode) Start() error                   { return n.plan.Start() }
func (n *selectTopNode) Next() (bool, error)            { return n.plan.Next() }
func (n *selectTopNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
func (n *selectTopNode) Values() map[string]interface{} { return n.plan.Values() }
func (n *selectTopNode) Close()                         { n.plan.Close() }

type renderInfo struct {
	numResults int
	fields     []*base.FieldDescription
	aliases    []string
}

type selectNode struct {
	p *Planner

	// main data source for the select node.
	source planNode

	// cache information about the original data source
	// collection name, meta-data, etc.
	sourceInfo sourceInfo

	// data related to rendering
	renderInfo *renderInfo

	// internal doc pointer
	// produced when Values()
	// is called.
	doc map[string]interface{}

	// top level filter expression
	// filter is split between select, scan, and typeIndexJoin.
	// The filters which only apply to the main collection
	// are stored in the root scanNode.
	// The filters that are defined on the root query, but apply
	// to the sub type are defined here in the select.
	// The filters that are defined on the subtype query
	// are defined in the subtype scan node.
	filter *parser.Filter

	// @todo restructure renderNode -> render, which is its own
	// object, and not a planNode.
}

func (n *selectNode) Init() error {
	return n.source.Init()
}

func (n *selectNode) Start() error {
	return n.source.Start()
}

// Next iterates through the source plan
// until a doc is returned, applies any
// remaining top level filtering, and
// renders the doc.
func (n *selectNode) Next() (bool, error) {
	for {
		if next, err := n.source.Next(); !next {
			return false, err
		}

		n.doc = n.source.Values()
		passes, err := parser.RunFilter(n.doc, n.filter, n.p.evalCtx)
		if err != nil {
			return false, err
		}

		if passes {
			n.renderDoc()
			return true, err
			// err :=
			// return err == nil, err
		}
		// didn't pass, keep looping
	}
}

// applies all the necessary rendering to doc
// as defined by the query statement. This includes
// aliases, and any transformations.
// Takes a doc map, and applies the necessary rendering.
// It also holds all the necessary render meta-data
// and ast parser data.
func (n *selectNode) renderDoc() error {
	renderData := map[string]interface{}{
		"numResults": n.renderInfo.numResults,
		"fields":     n.renderInfo.fields,
		"aliases":    n.renderInfo.aliases,
	}
	n.doc["__render"] = renderData
	return nil
}

func (n *selectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *selectNode) Values() map[string]interface{} {
	return n.doc
}

func (n *selectNode) Close() {
	n.source.Close()
}

// initSource is the main workhorse for recursively constructing
// all the necessary data source objects. This includes
// creating scanNodes, typeIndexJoinNodes, and splitting
// the necessary filters. Its designed to work with the
// planner.Select construction call.
func (n *selectNode) initSource(parsed *parser.Select) error {
	collectionName := parsed.Name
	sourcePlan, err := n.p.getSource(collectionName)
	if err != nil {
		return err
	}
	n.source = sourcePlan.plan
	n.sourceInfo = sourcePlan.info

	// split filter
	// apply the root filter to the source
	// and rootSubType filters to the selectNode
	// @todo: simulate splitting for now
	origScan, ok := n.source.(*scanNode)
	if ok {
		origScan.filter = n.filter
		n.filter = nil
	}

	n.renderInfo.numResults = 0
	// iterate looking just for fields
	// iterate again  just for Selects
	for _, field := range parsed.Fields {
		switch node := field.(type) {
		case *parser.Select:
			// continue //ignore for now
			// future:
			// plan := n.p.Select(node)
			// n.source := p.SubTypeIndexJoin(origScan, plan)
			f, found := n.sourceInfo.collectionDescription.GetField(node.GetName())
			if found {
				n.renderInfo.fields = append(n.renderInfo.fields, &f)
			}
		case *parser.Field:
			f, found := n.sourceInfo.collectionDescription.GetField(node.GetName())
			if found {
				n.renderInfo.fields = append(n.renderInfo.fields, &f)
			}
		}
		n.renderInfo.aliases = append(n.renderInfo.aliases, field.GetAlias())
		n.renderInfo.numResults++
	}

	return nil
}

// func (n *selectNode) initRender(fields []*base.FieldDescription, aliases []string) error {
// 	return n.p.render(fields, aliases)
// }

// Select constructs a SelectPlan
func (p *Planner) Select(parsed *parser.Select) (planNode, error) {
	s := &selectNode{p: p}
	s.filter = parsed.Filter
	limit := parsed.Limit  // ignore for now
	sort := parsed.OrderBy // ignore for now

	s.renderInfo = &renderInfo{}
	err := s.initSource(parsed)
	if err != nil {
		return nil, err
	}

	limitPlan, err := p.Limit(limit)
	if err != nil {
		return nil, err
	}

	sortPlan, err := p.OrderBy(sort)
	if err != nil {
		return nil, err
	}

	top := &selectTopNode{
		source: s,
		render: p.render(),
		limit:  limitPlan,
		sort:   sortPlan,
	}
	return top, nil
}
