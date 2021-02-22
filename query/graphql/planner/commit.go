package planner

import (
	"errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type commitSelectNode struct {
	p *Planner

	source *dagScanNode

	subRenderInfo map[string]renderInfo

	doc map[string]interface{}
}

func (n *commitSelectNode) Init() error {
	return n.source.Init()
}

func (n *commitSelectNode) Start() error {
	return n.source.Start()
}

func (n *commitSelectNode) Next() (bool, error) {
	if next, err := n.source.Next(); !next {
		return false, err
	}

	n.doc = n.source.Values()
	n.renderDoc()
	return true, nil
}

func (n *commitSelectNode) Values() map[string]interface{} {
	return n.doc
}

func (n *commitSelectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *commitSelectNode) Close() {
	n.source.Close()
}

func (n *commitSelectNode) Source() planNode {
	return n.source
}

func (p *Planner) CommitSelect(parsed *parser.CommitSelect) (planNode, error) {
	// check type of commit select (all, latest, one)
	var commit *commitSelectNode
	var err error
	switch parsed.Type {
	case parser.LatestCommits:
		commit, err = p.commitSelectLatest(parsed)
	}
	if err != nil {
		return nil, err
	}
	err = commit.initFields(parsed)
	if err != nil {
		return nil, err
	}
	slct := parsed.ToSelect()
	return p.SelectFromSource(slct, commit, false)
}

// commitSelectLatest is a CommitSelect node initalized with a headsetScanNode and a DocKey
func (p *Planner) commitSelectLatest(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	headset := p.HeadScan()
	if parsed.DocKey == "" {
		return nil, errors.New("Latest Commit query needs a DocKey")
	}
	// var field string
	// @todo: Get Collection field ID
	if parsed.FieldName == "" {
		parsed.FieldName = "C" // C for composite DAG
	}
	key := core.NewKey(parsed.DocKey + "/" + parsed.FieldName)
	headset.key = key
	dag.headset = headset
	dag.key = &key
	commit := &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}

	return commit, nil
}

// renderDoc applies the render meta-data to the
// links/previous sub selections for a commit type
// query.
func (n *commitSelectNode) renderDoc() {
	for subfield, info := range n.subRenderInfo {
		renderData := map[string]interface{}{
			"numResults": info.numResults,
			"fields":     info.fields,
			"aliases":    info.aliases,
		}
		for _, subcommit := range n.doc[subfield].([]map[string]interface{}) {
			subcommit["__render"] = renderData
		}

	}
}

func (n *commitSelectNode) initFields(parsed parser.Selection) error {
	for _, selection := range parsed.GetSelections() {
		switch node := selection.(type) {
		case *parser.Select:
			info := renderInfo{}
			for _, f := range node.Fields {
				info.fields = append(info.fields, f.GetName())
				info.aliases = append(info.aliases, f.GetAlias())
				info.numResults++
			}
			n.subRenderInfo[node.Name] = info
		}
	}
	return nil
}
