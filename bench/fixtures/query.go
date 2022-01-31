package fixtures

import (
	"github.com/sourcenetwork/defradb/db/base"
)

var (
	queryTemplate = `
	{{define "renderFields"}}
		{{range .Schema.Fields}} 
			{{if eq .Meta 0}} 
				{{.Name}} {{/* list all non meta fields, this includes related types and virtual fields */}}
			{{end}} 
		{{end}}


	{{end}}

	query {
		{{.Schema.Name }} {{if (HasArgs )}} ( {{.Args}} ) {{end}} {
			{{range .Schema.Fields}} 
				{{if eq .Meta 0}} 
					{{.Name}} {{/* list all non meta fields, this includes related types and virtual fields */}}
				{{end}} 
			{{end}}
		}
	}
	`

	argumentTemplate = `
	
	`
)

type queryTemplateContext struct {
	Schema  base.SchemaDescription
	HasArgs bool
	Args    string
}

// func QueryStringFromSchema(schema base.SchemaDescription, args map[string]interface{}) (string, error) {
// 	hasArgs := len(args) != 0
// 	argString := collectionArgString(args)
// 	tctx := queryTemplateContext{
// 		Schema:  schema,
// 		HasArgs: hasArgs,
// 		Args:    argString,
// 	}
// 	funcMap := template.FuncMap{
// 		"ToLower": strings.ToLower,
// 	}
// 	t, err := template.New("query").Funcs(funcMap).Parse(queryTemplate)
// 	if err != nil {
// 		return "", err
// 	}
// 	buf := new(bytes.Buffer)
// 	err = t.Execute(buf, tctx)

// 	return string(buf.Bytes()), err
// }

// argument here is the list of values to render into the GraphQL query.
// This applies to *all* arguments of the top level query field, as well
// as any child fields, independant of depth.
//
// Top-level query arguments are given as key-value where both are strings.
// Child query arguments are given as key-value where key is a string, and
// value is another map[string]interface{}, which itself may contain either
// query arguments, or more child queries.
//
// Eg: The following query
// query {
// 	user(filter: {age: {_gt: 10}}, limit: 10, sort: {age: DESC}) {
// 		_key
// 		name
// 		age

// 		friends(sort: {name: ASC}) {
// 			name
// 			points
// 		}
// 	}
// }
//
// Would have the following arg map:
// {
//	"filter": "{age: {_gt: 10}}",
//	"limit": "10",
//	"sort": "{age: DESC}",
//	"friends": map[string]interface{}{
//		"sort": "{name: ASC}"
//	}
// }
func collectionArgString(field string, args map[string]interface{}) map[string]string {
	// check for embedded
	// currentArgumnts := make([]string, 0)
	// argsByField := make(map[string]string)
	// for k, v := range args {
	// 	switch v.(type) {
	// 	case string:
	// 		formattedArg := fmt.Sprintf("%s: %s", k, v)
	// 		currentArgumnts = append(currentArgumnts, formattedArg)
	// 	case map[string]interface{}:
	// 		subArgs
	// 	}
	// }
	return nil
}
