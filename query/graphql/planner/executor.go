package planner

import (
	"errors"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	//github.com/uber-go/multierr
	gql "github.com/graphql-go/graphql"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Query is an external hook into the planNode
// system. It allows outside packages to
// execute and manage a query plan graph directly.
// Instead of using one of the available functions
// like ExecQuery(...).
// Currently, this is used by the collection.Update
// system.
type Query planNode

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

func (e *QueryExecutor) MakeSelectQuery(db client.DB, txn client.Txn, selectStmt *parser.Select) (Query, error) {
	if selectStmt == nil {
		return nil, errors.New("Cannot create query without a selection")
	}
	planner := makePlanner(db, txn)
	return planner.makePlan(selectStmt)
}

func (e *QueryExecutor) ExecQuery(db client.DB, txn client.Txn, query string, args ...interface{}) ([]map[string]interface{}, error) {
	q, err := e.parseQueryString(query)
	if err != nil {
		return nil, err
	}

	planner := makePlanner(db, txn)
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
