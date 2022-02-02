package fixtures

var (
	gTypeToGQLType = map[string]string{
		"int":     "Int",
		"string":  "String",
		"float64": "Float",
		"float32": "Float",
		"bool":    "Boolean",
	}
)

type User struct {
	Name     string `faker:"name"`
	Age      int
	Points   float32 `faker:"amount"`
	Verified bool
}
