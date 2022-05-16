package connor

type FilterKey interface {
	GetProp(data interface{}) interface{}
	GetOperatorOrDefault(defaultOp string) string
}
