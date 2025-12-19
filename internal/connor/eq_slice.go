package connor

import (
	"cmp"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor/numbers"
)

var (
	ErrSliceTypeNotFound = errors.New("slice type not found")
)

func equalAnyToAnySlice(a any, b []any) (bool, error) {
	switch aTyped := a.(type) {
	case []bool:
		return equalSlice(aTyped, b), nil
	case []int:
		return equalSliceNumeric(aTyped, b), nil
	case []int32:
		return equalSliceNumeric(aTyped, b), nil
	case []int64:
		return equalSliceNumeric(aTyped, b), nil
	case []float32:
		return equalSliceNumeric(aTyped, b), nil
	case []float64:
		return equalSliceNumeric(aTyped, b), nil
	case []string:
		return equalSlice(aTyped, b), nil
	case []time.Time:
		return equalSlice(aTyped, b), nil

	case []immutable.Option[bool]:
		return equalOptionSlice(aTyped, b), nil
	case []immutable.Option[int]:
		return equalOptionSliceNumeric(aTyped, b), nil
	case []immutable.Option[int32]:
		return equalOptionSliceNumeric(aTyped, b), nil
	case []immutable.Option[int64]:
		return equalOptionSliceNumeric(aTyped, b), nil
	case []immutable.Option[float32]:
		return equalOptionSliceNumeric(aTyped, b), nil
	case []immutable.Option[float64]:
		return equalOptionSliceNumeric(aTyped, b), nil
	case []immutable.Option[string]:
		return equalOptionSlice(aTyped, b), nil
	case []immutable.Option[time.Time]:
		return equalOptionSlice(aTyped, b), nil
	default:
		return false, ErrSliceTypeNotFound
	}
}

// This is basically the most effecient approach to slice comparison
// since it avoids 1) reflection 2) allocation
func equalSlice[T comparable](a []T, b any) bool {
	switch bTyped := b.(type) {
	case []T:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			if v != bTyped[i] {
				return false
			}
		}
		return true
	case []any:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			if bv, ok := bTyped[i].(T); !ok || v != bv {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// This is basically the most effecient approach to slice comparison
// since it avoids 1) reflection 2) allocation
func equalSliceNumeric[T cmp.Ordered](a []T, b any) bool {
	switch bTyped := b.(type) {
	case []T:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			if v != bTyped[i] {
				return false
			}
		}
		return true
	case []any:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			bv, ok := bTyped[i].(T)
			if !ok && !numbers.Equal(v, bTyped[i]) {
				return false
			} else if ok && v != bv {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func equalOption[T comparable](a, b immutable.Option[T]) bool {
	if !a.HasValue() && !b.HasValue() {
		return true
	}
	if a.HasValue() != b.HasValue() {
		return false
	}
	return a.Value() == b.Value()
}

func equalOptionSlice[T comparable](a []immutable.Option[T], b any) bool {
	switch bTyped := b.(type) {
	case []immutable.Option[T]:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			if !equalOption(v, bTyped[i]) {
				return false
			}
		}
		return true
	case []any:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			hasVal := v.HasValue()
			if !hasVal && bTyped[i] == nil {
				continue
			} else if hasVal && bTyped[i] == v.Value() {
				continue
			}
			return false
		}
		return true
	default:
		return false
	}
}

func equalOptionSliceNumeric[T cmp.Ordered](a []immutable.Option[T], b any) bool {
	switch bTyped := b.(type) {
	case []immutable.Option[T]:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			if !equalOption(v, bTyped[i]) {
				return false
			}
		}
		return true
	case []any:
		if len(a) != len(bTyped) {
			return false
		}
		for i, v := range a {
			hasVal := v.HasValue()
			if !hasVal && bTyped[i] == nil {
				continue
			} else if hasVal && numbers.Equal(bTyped[i], v.Value()) {
				continue
			}
			return false
		}
		return true
	default:
		return false
	}
}
