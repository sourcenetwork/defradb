package fixtures

import (
	"bytes"
	"context"
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

func (ctx Context) GenerateFixtureDocs() ([]string, error) {
	results := make([]string, len(ctx.types))
	for i, t := range ctx.types {
		val := reflect.New(reflect.TypeOf(t))
		err := faker.FakeData(val)
		if err != nil {
			return nil, err
		}

	}
}

// extractGQLFromType extracts a GraphQL SDL definition as a string
// from a given type struct
func extractGQLFromType(t interface{}) (string, error) {
	var buf *bytes.Buffer
	if reflect.TypeOf(t).Kind() != reflect.Struct {
		return "", errors.New("given type is not a struct")
	}

	// get name
	tt := reflect.TypeOf(t)
	name := tt.Name()

	// write the GQL SDL object to the buffer, field by field
	fmt.Fprintf(buf, "type %s {", name)
	for i := 0; i < tt.NumField(); i++ {
		f := tt.Field(i)
		fname := f.Name
		ftype := f.Type.Name()
		fmt.Fprintf(buf, "\t%s: %s\n", fname, ftype)
	}
	fmt.Fprint(buf, "}")

	return buf.String(), nil
}
