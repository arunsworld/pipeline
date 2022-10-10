package pipeline

type preFilter[T any] struct {
	allowFunc func(T) bool
}

func (p preFilter[T]) init(v pipeline[T]) pipeline[T] {
	// v.preFilters = append(v.preFilter, newFilterOperation(p.allowFunc))
	return v
}

func newFilterOperation[T any](allowFunc func(T) bool) FilterOperation[T] {
	return filterOperation[T]{
		allowFunc: allowFunc,
	}
}

type filterOperation[T any] struct {
	allowFunc func(T) bool
}

func (fo filterOperation[T]) Allow(v T) bool {
	return fo.allowFunc(v)
}

func chainFilters[T any](filters []FilterOperation[T]) FilterOperation[T] {
	return newFilterOperation(func(v T) bool {
		for _, filter := range filters {
			if !filter.Allow(v) {
				return false
			}
		}
		return true
	})
}

func noopFilter[T any]() FilterOperation[T] {
	return newFilterOperation(func(v T) bool {
		return true
	})
}
