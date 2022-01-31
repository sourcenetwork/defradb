package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/bxcodec/faker"
)

var (
	registeredFixtures = map[string][]interface{}{
		"user_simple": {User{}},
	}
)

type Context struct {
	ctx context.Context

	schema string
	types  []interface{}
}

func WithSchema(ctx context.Context, schemaName string) Context {
	return Context{
		ctx:    ctx,
		schema: schemaName,
		types:  registeredFixtures[schemaName],
	}
}

// Types returns the defined types for this fixture set
func (ctx Context) Types() []interface{} {
	return ctx.types
}

// Type returns type at the given index in the fixture set
func (ctx Context) Type(index int) interface{} {
	return ctx.types[index]
}

// TypeName returns the name of the type at the given index
// in the fixture set
func (ctx Context) TypeName(index int) string {
	return reflect.TypeOf(ctx.types[index]).Name()
}

// GenerateFixtureDocs uses the faker fixture system to
// randomly generate a new set of documents matching the defined
// struct types within the context.
func (ctx Context) GenerateDocs() ([]string, error) {
	results := make([]string, len(ctx.types))
	for i, t := range ctx.types {
		val := reflect.New(reflect.TypeOf(t)).Interface()

		// generate our new random struct and
		// write it to our reflected variable
		if err := faker.FakeData(val); err != nil {
			return nil, err
		}

		buf, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		results[i] = string(buf)
	}

	return results, nil
}

// func  GenerateFullQueryString() ([]string, error) {
// 	// q := ""
// 	queries := make([]string, len(ctx.types))
// 	for i, t := range ctx.types {
// 		q, err := queryStringFromSchema(ctx.schema, nil)
// 	}
// 	return nil, nil
// }

// extractGQLFromType extracts a GraphQL SDL definition as a string
// from a given type struct
func ExtractGQLFromType(t interface{}) (string, error) {
	var buf bytes.Buffer

	if reflect.TypeOf(t).Kind() != reflect.Struct {
		return "", errors.New("given type is not a struct")
	}

	// get name
	tt := reflect.TypeOf(t)
	name := tt.Name()

	// write the GQL SDL object to the buffer, field by field
	fmt.Fprintf(&buf, "type %s {\n", name)
	for i := 0; i < tt.NumField(); i++ {
		// @todo: Handle non-scalar types
		f := tt.Field(i)
		fname := f.Name
		ftype := f.Type.Name()
		gqlType := gTypeToGQLType[ftype]
		fmt.Fprintf(&buf, "\t%s: %s\n", fname, gqlType)
	}
	fmt.Fprint(&buf, "}")

	return buf.String(), nil
}
