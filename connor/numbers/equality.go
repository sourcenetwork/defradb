package numbers

func Equal(condition, data interface{}) bool {
	uc := TryUpcast(condition)
	ud := TryUpcast(data)

	switch ucv := uc.(type) {
	case int64:
		if udv, ok := ud.(int64); ok {
			return ucv == udv
		} else if udv, ok := ud.(float64); ok {
			fucv := float64(ucv)
			return fucv == udv && int64(fucv) == ucv
		}
		return false
	case float64:
		if udv, ok := ud.(float64); ok {
			return ucv == udv
		} else if udv, ok := ud.(int64); ok {
			iucv := int64(ucv)
			return iucv == udv && float64(iucv) == ucv
		}
		return false
	default:
		return false
	}
}
