package pipeline

// context is deliberately left out - instead the individual filter/transform functions are expected to be context aware if really needed
// streaming ends when input channel is closed - until then it blocks
type Pipeline[T any] interface {
	Apply([]T) ([]T, error)
	ApplyAndFold([]T, FoldOperation[T]) (T, error)
	Stream(<-chan T, chan<- T) error
}

type Option[T any] interface {
	init(pipeline[T]) pipeline[T]
}

func PreFilter[T any](allowFunc func(T) bool) Option[T] {
	return preFilter[T]{
		allowFunc: allowFunc,
	}
}

type Components[T any] struct {
	PreFilters   []FilterOperation[T]
	Transformers []TransformOperation[T]
	PostFilters  []FilterOperation[T]
	Concurrency  int
}

func New[T any](components Components[T]) Pipeline[T] {
	return newPipeline(components)
}

// operation that evaluates whether an element should stay or not
type FilterOperation[T any] interface {
	Allow(T) bool
}

// filterFunc function - returning true causes an element to stay; false causes it to be discarded
// function should be thread-safe
func NewFilterOperation[T any](allowFunc func(T) bool) FilterOperation[T] {
	return newFilterOperation(allowFunc)
}

// operations that transforms an element to another
// alternatively may ask to be filtered out
// or return a terminal error
type TransformOperation[T any] interface {
	Transform(T) (T, bool, error)
}

// transformFunc should handle transformation, judgement on filtration and error
// function should be threadsafe
func NewTransformAndFilterOperation[T any](transformFunc func(T) (T, bool, error)) TransformOperation[T] {
	return newTransformAndFilterOperation(transformFunc)
}

// transformFunc should handle transformation and judgement on filtration
func NewMustTransformAndFilterOperation[T any](transformFunc func(T) (T, bool)) TransformOperation[T] {
	return newMustTransformAndFilterOperation(transformFunc)
}

// transformFunc should handle transformation error - all elements are passed through
func NewTransformOperation[T any](transformFunc func(T) (T, error)) TransformOperation[T] {
	return newTransformOperation(transformFunc)
}

// transformFunc should handle transformation - no error should be possible; all elements are passed through
func NewMustTransformOperation[T any](transformFunc func(T) T) TransformOperation[T] {
	return newMustTransformOperation(transformFunc)
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
