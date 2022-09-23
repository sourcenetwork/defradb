package numbers

func TryUpcast(n any) any {
	switch nn := n.(type) {
	case int8:
		return int64(nn)
	case int16:
		return int64(nn)
	case int32:
		return int64(nn)
	case int:
		return int64(nn)
	case int64:
		return nn
	case uint64:
		nnn := int64(nn)
		if uint64(nnn) == nn { // if we can safely convert from uint64 -> int64 without losing data
			return nnn
		}
		return n
	case float32:
		return float64(nn)
	case float64:
		return nn
	default:
		return n
	}
}
