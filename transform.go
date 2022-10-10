package pipeline

func newTransformAndFilterOperation[T any](transformFunc func(T) (T, bool, error)) TransformOperation[T] {
	return transformOperation[T]{
		transformFunc: transformFunc,
	}
}

func newMustTransformAndFilterOperation[T any](transformFunc func(T) (T, bool)) TransformOperation[T] {
	return transformOperation[T]{
		transformFunc: func(v T) (T, bool, error) {
			v, ok := transformFunc(v)
			return v, ok, nil
		},
	}
}

func newTransformOperation[T any](transformFunc func(T) (T, error)) TransformOperation[T] {
	return transformOperation[T]{
		transformFunc: func(v T) (T, bool, error) {
			v, err := transformFunc(v)
			return v, true, err
		},
	}
}

func newMustTransformOperation[T any](transformFunc func(T) T) TransformOperation[T] {
	return transformOperation[T]{
		transformFunc: func(v T) (T, bool, error) {
			v = transformFunc(v)
			return v, true, nil
		},
	}
}

type transformOperation[T any] struct {
	transformFunc func(T) (T, bool, error)
}

func (to transformOperation[T]) Transform(v T) (T, bool, error) {
	return to.transformFunc(v)
}

func chainedTransformers[T any](transformers []TransformOperation[T]) TransformOperation[T] {
	return transformOperation[T]{
		transformFunc: func(v T) (T, bool, error) {
			for _, transformer := range transformers {
				newv, ok, err := transformer.Transform(v)
				if err != nil {
					return v, ok, err
				}
				if !ok {
					return v, false, nil
				}
				v = newv
			}
			return v, true, nil
		},
	}
}

func noopTransformer[T any]() TransformOperation[T] {
	return newMustTransformOperation(func(v T) T {
		return v
	})
}
