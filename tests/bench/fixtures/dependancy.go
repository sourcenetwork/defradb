package fixtures

import (
	"fmt"
	"reflect"
)

// return dependants of a type
//
// one-to-one: A->B | A (primary) depends on B (secondary)
// one-to-many: A->[]B | A (one) depends on []B (many)
func dependants(val reflect.Type) ([]*Node, error) {
	deps := make([]*Node, 0)

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		if !isRelationDependantField(f.Type) {
			continue
		}

		// get fixture tag
		// if its empty, skip
		tag := f.Tag.Get(fixtureTag)
		tg, err := parseTag(tag)
		if err != nil {
			return nil, err
		}

		switch tg.rel {
		case oneToOne:
			// the second entry in the tags comma seperated list
			// has to be primary or empty
			if !tg.isPrimary {
				// secondary, no dependancy
				// next field
				continue
			}

			// primary, dependancy
			deps = append(deps, NewNode(f.Type.Elem().Name(), f.Type.Elem()))
		case oneToMany:
			deps = append(deps, NewNode(f.Type.Elem().Name(), f.Type.Elem()))

		default:
			return nil, fmt.Errorf("invalid 'fixture' tag format, missing relation type: %s", tag)
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
func isRelationDependantField(field reflect.Type) bool {
	return isStructPointer(field) || isSliceStructPointer(field)
}

func isStructPointer(t reflect.Type) bool {
	return (t.Kind() == reflect.Pointer &&
		t.Elem().Kind() == reflect.Struct)
}

func isSliceStructPointer(t reflect.Type) bool {
	return (t.Kind() == reflect.Slice && isStructPointer(t.Elem()))
}
