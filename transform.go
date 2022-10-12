package pipeline

type transformer[T any] struct {
	transformFuncWithFilter     func(T) (T, bool, error)
	mustTransformFuncWithFilter func(T) (T, bool)
	transformFunc               func(T) (T, error)
	mustTransformFunc           func(T) T
}

func (t transformer[T]) init(v *pipeline[T]) {
	switch {
	case t.transformFuncWithFilter != nil:
		v.addTransformFunc(t.transformFuncWithFilter)
	case t.mustTransformFuncWithFilter != nil:
		v.addTransformFunc(func(v T) (T, bool, error) {
			newv, ok := t.mustTransformFuncWithFilter(v)
			return newv, ok, nil
		})
	case t.transformFunc != nil:
		v.addTransformFunc(func(v T) (T, bool, error) {
			newv, err := t.transformFunc(v)
			return newv, true, err
		})
	case t.mustTransformFunc != nil:
		v.addTransformFunc(func(v T) (T, bool, error) {
			newv := t.mustTransformFunc(v)
			return newv, true, nil
		})
	default:
	}
}

func noopTransformFunc[T any](v T) (T, bool, error) {
	return v, true, nil
}
