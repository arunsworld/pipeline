package pipeline

type Pipeline[T any] interface {
	PipelineSetup[T]
	PiplineExecute[T]
}

type PipelineSetup[T any] interface {
	Concurrency(int) Pipeline[T]
	Prefilter(func(T) bool) Pipeline[T]
	Postfilter(func(T) bool) Pipeline[T]
	TransformWithFilter(func(T) (T, bool, error)) Pipeline[T]
	MustTransformWithFilter(func(T) (T, bool)) Pipeline[T]
	Transform(func(T) (T, error)) Pipeline[T]
	MustTransform(func(T) T) Pipeline[T]
}

// context is deliberately left out - instead the individual filter/transform functions are expected to be context aware if really needed
// streaming ends when input channel is closed - until then it blocks
type PiplineExecute[T any] interface {
	Apply([]T) ([]T, error)
	ApplyAndFold([]T, FoldOperation[T]) (T, error)
	Stream(<-chan T, chan<- T) error
}

func New[T any]() Pipeline[T] {
	return newPipeline[T]()
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
