package pipeline

type preFilter[T any] struct {
	allowFunc func(T) bool
}

func (p preFilter[T]) init(v *pipeline[T]) {
	v.addPreFilter(p.allowFunc)
}

type postFilter[T any] struct {
	allowFunc func(T) bool
}

func (p postFilter[T]) init(v *pipeline[T]) {
	v.addPostFilter(p.allowFunc)
}

func noopAllow[T any](T) bool { return true }

type concurrency[T any] struct {
	concurrency int
}

func (c concurrency[T]) init(v *pipeline[T]) {
	v.concurrency = c.concurrency
}
