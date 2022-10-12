package pipeline

// context is deliberately left out - instead the individual filter/transform functions are expected to be context aware if really needed
// streaming ends when input channel is closed - until then it blocks
type Pipeline[T any] interface {
	Apply([]T) ([]T, error)
	ApplyAndFold([]T, FoldOperation[T]) (T, error)
	Stream(<-chan T, chan<- T) error
}

type Option[T any] interface {
	init(*pipeline[T])
}

func New[T any](options ...Option[T]) Pipeline[T] {
	return newPipeline(options)
}

// Options
func Concurrency[T any](c int) Option[T] {
	return concurrency[T]{
		concurrency: c,
	}
}

func PreFilter[T any](allowFunc func(T) bool) Option[T] {
	return preFilter[T]{
		allowFunc: allowFunc,
	}
}

func PostFilter[T any](allowFunc func(T) bool) Option[T] {
	return postFilter[T]{
		allowFunc: allowFunc,
	}
}

func TransformWithFilter[T any](transformFunc func(T) (T, bool, error)) Option[T] {
	return transformer[T]{
		transformFuncWithFilter: transformFunc,
	}
}

func MustTransformWithFilter[T any](transformFunc func(T) (T, bool)) Option[T] {
	return transformer[T]{
		mustTransformFuncWithFilter: transformFunc,
	}
}

func Transform[T any](transformFunc func(T) (T, error)) Option[T] {
	return transformer[T]{
		transformFunc: transformFunc,
	}
}

func MustTransform[T any](transformFunc func(T) T) Option[T] {
	return transformer[T]{
		mustTransformFunc: transformFunc,
	}
}

// folds two T elements into one
type FoldOperation[T any] interface {
	Fold(T, T) (T, error)
}

func NewFoldOperation[T any](foldFunc func(T, T) (T, error)) FoldOperation[T] {
	return foldOperation[T]{
		foldFunc: foldFunc,
	}
}

func NewMustFoldOperation[T any](foldFunc func(T, T) T) FoldOperation[T] {
	return foldOperation[T]{
		foldFunc: func(v1, v2 T) (T, error) {
			return foldFunc(v1, v2), nil
		},
	}
}

type foldOperation[T any] struct {
	foldFunc func(T, T) (T, error)
}

func (fo foldOperation[T]) Fold(v1, v2 T) (T, error) {
	return fo.foldFunc(v1, v2)
}
