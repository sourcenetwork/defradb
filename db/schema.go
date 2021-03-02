package db

import (
	"fmt"

	"github.com/sourcenetwork/defradb/query/graphql/schema"

	gql "github.com/graphql-go/graphql"
)

// LoadSchema takes the provided schema in SDL format, and applies it to the database,
// and creates the necessary collections, query types, etc.
func (db *DB) LoadSchema(schema string) error {
	// @todo: create collection after generating query types
	types, err := db.schema.Generator.FromSDL(schema)
	if err != nil {
		return err
	}
	colDesc, err := db.schema.Generator.CreateDescriptions(types)
	if err != nil {
		return err
	}
	for _, desc := range colDesc {
		// fmt.Println(desc)
		if _, err := db.CreateCollection(desc); err != nil {
			return err
		}
	}

	user := db.schema.Schema().Type("users")
	if user != nil {
		if user.Error() != nil {
			fmt.Println("User type error:", user.Error())
		} else {
			fmt.Println("User Type:", user, "fields:", user.(*gql.Object).Fields())
		}
	}
	return nil
}

func (db *DB) LoadSchemaIfNotExists(schema string) error { return nil }

func (db *DB) SchemaManager() *schema.SchemaManager {
	return db.schema
}
