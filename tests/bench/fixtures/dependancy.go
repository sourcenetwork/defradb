package fixtures

import (
	"fmt"
	"reflect"
	"strings"
)

// return dependants of a type
//
// one-to-one: A->B | A (primary) depends on B (secondary)
// one-to-many: A->[]B | A (one) depends on []B (many)
func dependants(val reflect.Type) ([]string, error) {
	deps := make([]string, 0)

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		fmt.Printf("%s field %s => %s\n", val.Name(), f.Name, f.Type.Name())
		if !isRelationDependantField(f) {
			continue
		}

		// get fixture tag
		// if its empty, skip
		tag := f.Tag.Get(fixtureTag)
		if tag == "" {
			return []string{}, fmt.Errorf("field %s is missing 'fixture' tag", f.Name)
		}

		tags := strings.Split(tag, ",")
		switch tags[0] {
		case string(oneToOne):
			// the second entry in the tags comma seperated list
			// has to be primary or empty
			if len(tags) < 2 || tags[1] != "primary" {
				// secondary, no dependancy
				// next field
				continue
			}

			// primary, dependancy
			deps = append(deps, f.Type.Elem().Name())
		case string(oneToMany):
			deps = append(deps, f.Type.Elem().Name())

		default:
			return []string{}, fmt.Errorf("invalid 'fixture' tag format, missing relation type: %s", tag)
		}
	}

	return deps, nil
}

// isRelationDependantField returns whether or not the given
// type is a field that contains a relation
//
// A field contains a relation iff:
// 1) It is a pointer to a struct
// or
// 2) It is a slice of struct pointers
func isRelationDependantField(field reflect.StructField) bool {
	t := field.Type
	for {
		switch t.Kind() {
		case reflect.Struct:
			return true
		case reflect.Pointer:
			t = t.Elem()
		default:
			return false
		}
	}
}

func isStructPointer(t reflect.Type) bool {
	return (t.Kind() == reflect.Pointer &&
		t.Elem().Kind() == reflect.Struct)
}

func isSliceStructPointer(t reflect.Type) bool {
	return (t.Kind() == reflect.Slice && isStructPointer(t.Elem()))
}
