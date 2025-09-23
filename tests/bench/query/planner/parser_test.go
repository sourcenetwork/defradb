package planner

import (
	"testing"

	gast "github.com/sourcenetwork/graphql-go/language/ast"
	gparser "github.com/sourcenetwork/graphql-go/language/parser"
	source "github.com/sourcenetwork/graphql-go/language/source"
	//vast "github.com/vektah/gqlparser/v2/ast"
	// vparser "github.com/vektah/gqlparser/v2/parser"
)

// var vektahParserDoc *vast.QueryDocument
var gqlgParserDoc *gast.Document

var query = `
		subscription {
			User(docId: "bae-9d94d475-045a-53bc-a3cd-eabe880836ad", filter: {name: {_gt: 10}, points: {_in: [1,2,3,4,5,6]}}) {
				_docID
				_version {
					cid
					delta
					links {
						cid
						name
					}
				}

				name
				points
				age
				test
			}
		}
	`

// KEEPING THIS AS IS, commented out to avoid the dependency addition
/*
func Benchmark_RawParser_vektah_UserSimple_ParseQuery(b *testing.B) {
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vektahParserDoc, err = vparser.ParseQuery(&vast.Source{
			Name:  "GraphQL",
			Input: query,
		})
		if err != nil {
			b.Error(err)
			return
		}
	}
}
*/

func Benchmark_RawParser_graphqlgo_UserSimple_ParseQuery(b *testing.B) {
	query := `
		subscription {
			User(docId: "bae-9d94d475-045a-53bc-a3cd-eabe880836ad", filter: {name: {_gt: 10}, points: {_in: [1,2,3,4,5,6]}}) {
				_docID
				_version {
					cid
					delta
					links {
						cid
						name
					}
				}

				name
				points
				age
				test
			}
		}
	`

	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gqlgParserDoc, err = gparser.Parse(gparser.ParseParams{
			Source: &source.Source{
				Body: []byte(query),
				Name: "GraphQL",
			},
		})
		if err != nil {
			b.Error(err)
			return
		}
	}
}
