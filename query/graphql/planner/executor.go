package planner

import (
	"errors"
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	//github.com/uber-go/multierr
	gql "github.com/graphql-go/graphql"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

type QueryExecutor struct {
	// some context
	// schema manager
	SchemaManager *schema.SchemaManager
}

func NewQueryExecutor(manager *schema.SchemaManager) (*QueryExecutor, error) {
	// sm, err := schema.NewSchemaManager()
	// if err != nil {
	// 	return nil, nil
	// }
	if manager == nil {
		return nil, errors.New("SchemaManager cannot be nil")
	}

	// g := schema.NewGenerator(sm)
	return &QueryExecutor{
		SchemaManager: manager,
	}, nil
}

// func (e *QueryExecutor) ExecQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {

// }

func (e *QueryExecutor) ExecQuery(txn core.Txn, query string, args ...interface{}) ([]map[string]interface{}, error) {
	q, err := e.parseQueryString(query)
	if err != nil {
		return nil, err
	}

	planner := makePlanner(txn)
	return planner.queryDocs(q)
}

func (e *QueryExecutor) parseQueryString(query string) (*parser.Query, error) {
	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "GraphQL request",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	schema := e.SchemaManager.Schema()
	validationResult := gql.ValidateDocument(schema, doc, nil)
	if !validationResult.IsValid {
		return nil, fmt.Errorf("%v", validationResult.Errors)
	}

	return parser.ParseQuery(doc)
}
