package connor

type EvaluateFunc = func(condition, data interface{}) (bool, error)

var opMap = map[string]EvaluateFunc{}
