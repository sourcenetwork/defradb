package parser

import (
	"github.com/graphql-go/graphql/language/ast"
)

type Statement interface {
	GetStatement() ast.Node
}
