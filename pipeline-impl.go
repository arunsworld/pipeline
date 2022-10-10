package pipeline

import (
	"context"
	"sort"

	"github.com/arunsworld/nursery"
)

func newPipeline[T any](components Components[T]) Pipeline[T] {
	concurrency := 1
	if components.Concurrency > 1 {
		concurrency = components.Concurrency
	}
	var preFilter FilterOperation[T]
	if len(components.PreFilters) > 0 {
		preFilter = chainFilters(components.PreFilters)
	} else {
		preFilter = noopFilter[T]()
	}
	var transformer TransformOperation[T]
	if len(components.Transformers) > 0 {
		transformer = chainedTransformers(components.Transformers)
	} else {
		transformer = noopTransformer[T]()
	}
	var postFilter FilterOperation[T]
	if len(components.PostFilters) > 0 {
		postFilter = chainFilters(components.PostFilters)
	} else {
		postFilter = noopFilter[T]()
	}
	return pipeline[T]{
		concurrency: concurrency,
		preFilter:   preFilter,
		transformer: transformer,
		postFilter:  postFilter,
	}
}

type pipeline[T any] struct {
	concurrency int
	preFilter   FilterOperation[T]
	transformer TransformOperation[T]
	postFilter  FilterOperation[T]
}

func (p pipeline[T]) Apply(input []T) ([]T, error) {
	switch {
	case p.concurrency <= 1:
		return p.applySequential(input)
	default:
		return p.applyConcurrent(input)
	}
}

func (p pipeline[T]) ApplyAndFold(input []T, foldOp FoldOperation[T]) (T, error) {
	switch {
	case p.concurrency <= 1:
		return p.applyAndFoldSequential(input, foldOp)
	default:
		return p.applyAndFoldConcurrently(input, foldOp)
	}
}

func (p pipeline[T]) Stream(inCh <-chan T, outCh chan<- T) error {
	switch {
	case p.concurrency <= 1:
		return p.streamSequential(inCh, outCh)
	default:
		return p.streamConcurrent(inCh, outCh)
	}
}

func (p pipeline[T]) process(v T) (T, bool, error) {
	if !p.preFilter.Allow(v) {
		return v, false, nil
	}
	v, ok, err := p.transformer.Transform(v)
	if err != nil {
		return v, ok, err
	}
	if !ok {
		return v, false, nil
	}
	if !p.postFilter.Allow(v) {
		return v, false, nil
	}
	return v, true, nil
}

func (p pipeline[T]) applySequential(input []T) ([]T, error) {
	result := make([]T, 0, len(input))
	for _, v := range input {
		v, ok, err := p.process(v)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result = append(result, v)
	}
	return result, nil
}

type elementWithIndex[T any] struct {
	idx  int
	data T
}

func (p pipeline[T]) applyConcurrent(input []T) ([]T, error) {
	buffer := make([]elementWithIndex[T], 0, len(input))
	inCh := make(chan elementWithIndex[T], 3)
	outCh := make(chan elementWithIndex[T], 3)
	err := nursery.RunConcurrently(
		func(ctx context.Context, _ chan error) {
			defer close(inCh)
			for idx, v := range input {
				select {
				case inCh <- elementWithIndex[T]{idx: idx, data: v}:
				case <-ctx.Done():
					return
				}
			}
		},
		func(_ context.Context, errCh chan error) {
			if err := p.streamConcurrentElementsWithIndex(inCh, outCh); err != nil {
				errCh <- err
			}
			close(outCh)
		},
		func(context.Context, chan error) {
			for v := range outCh {
				buffer = append(buffer, v)
			}
		},
	)
	sort.Slice(buffer, func(i, j int) bool {
		return buffer[i].idx < buffer[j].idx
	})
	result := make([]T, 0, len(buffer))
	for _, v := range buffer {
		result = append(result, v.data)
	}
	return result, err
}

func (p pipeline[T]) streamSequential(inCh <-chan T, outCh chan<- T) error {
	for v := range inCh {
		v, ok, err := p.process(v)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		outCh <- v
	}
	return nil
}

func (p pipeline[T]) streamConcurrent(inCh <-chan T, outCh chan<- T) error {
	return nursery.RunMultipleCopiesConcurrently(p.concurrency,
		func(ctx context.Context, errCh chan error) {
			for {
				select {
				case v, ok := <-inCh:
					if !ok {
						return
					}
					v, ok, err := p.process(v)
					if err != nil {
						errCh <- err
						return
					}
					if !ok {
						break
					}
					outCh <- v
				case <-ctx.Done():
					return
				}
			}
		},
	)
}

func (p pipeline[T]) streamConcurrentElementsWithIndex(inCh <-chan elementWithIndex[T], outCh chan<- elementWithIndex[T]) error {
	return nursery.RunMultipleCopiesConcurrently(p.concurrency,
		func(ctx context.Context, errCh chan error) {
			for {
				select {
				case v, ok := <-inCh:
					if !ok {
						return
					}
					data, ok, err := p.process(v.data)
					if err != nil {
						errCh <- err
						return
					}
					if !ok {
						break
					}
					outCh <- elementWithIndex[T]{idx: v.idx, data: data}
				case <-ctx.Done():
					return
				}
			}
		},
	)
}

func (p pipeline[T]) applyAndFoldSequential(input []T, foldOp FoldOperation[T]) (T, error) {
	var result T
	if len(input) == 0 {
		return result, nil
	}
	nextIdx := 0
	for idx, v := range input {
		v, ok, err := p.process(v)
		if err != nil {
			return result, err
		}
		if !ok {
			continue
		}
		result = v
		nextIdx = idx + 1
		break
	}
	if nextIdx == 0 || nextIdx == len(input) {
		return result, nil
	}
	for i := nextIdx; i < len(input); i++ {
		v, ok, err := p.process(input[i])
		if err != nil {
			return result, err
		}
		if !ok {
			continue
		}
		nr, err := foldOp.Fold(result, v)
		if err != nil {
			return result, err
		}
		result = nr
	}
	return result, nil
}

func (p pipeline[T]) applyAndFoldConcurrently(input []T, foldOp FoldOperation[T]) (T, error) {
	var result T
	if len(input) == 0 {
		return result, nil
	}
	nextIdx := 0
	for idx, v := range input {
		v, ok, err := p.process(v)
		if err != nil {
			return result, err
		}
		if !ok {
			continue
		}
		result = v
		nextIdx = idx + 1
		break
	}
	if nextIdx == 0 || nextIdx == len(input) {
		return result, nil
	}

	inCh := make(chan T, 3)
	outCh := make(chan T, 3)
	err := nursery.RunConcurrently(
		func(ctx context.Context, _ chan error) {
			defer close(inCh)
			for i := nextIdx; i < len(input); i++ {
				select {
				case inCh <- input[i]:
				case <-ctx.Done():
					return
				}
			}
		},
		func(ctx context.Context, errCh chan error) {
			defer close(outCh)
			err := nursery.RunMultipleCopiesConcurrentlyWithContext(ctx, p.concurrency,
				func(ctx context.Context, errCh chan error) {
					var intermediateResult T
					haveIntermediateResult := false
					defer func() {
						if haveIntermediateResult {
							outCh <- intermediateResult
						}
					}()
					for {
						select {
						case v, ok := <-inCh:
							if !ok {
								return
							}
							v, ok, err := p.process(v)
							if err != nil {
								errCh <- err
								return
							}
							if !ok {
								break
							}
							if !haveIntermediateResult {
								haveIntermediateResult = true
								intermediateResult = v
							} else {
								_ir, err := foldOp.Fold(intermediateResult, v)
								if err != nil {
									errCh <- err
									return
								}
								intermediateResult = _ir
							}
						case <-ctx.Done():
							return
						}
					}
				},
			)
			if err != nil {
				errCh <- err
			}
		},
		func(_ context.Context, errCh chan error) {
			for v := range outCh {
				nr, err := foldOp.Fold(result, v)
				if err != nil {
					errCh <- err
					for range outCh {
					}
					return
				}
				result = nr
			}
		},
	)
	return result, err
}