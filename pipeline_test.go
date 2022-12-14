package pipeline_test

import (
	"reflect"
	"sort"
	"testing"

	p "github.com/arunsworld/pipeline"
)

func Test_Pipeline_With_Filters_And_Transformers(t *testing.T) {
	t.Run("can run", func(t *testing.T) {
		input := numbersFromAtoB(0, 100)
		expectedOutput := []int{7, 19, 67, 199, 259, 487, 679, 787, 1027, 1159, 1447, 1939, 2119, 2707, 3139, 3367, 3847, 4099, 4627, 5479, 5779, 6727, 7399, 7747, 8467, 8839, 9607}

		t.Run("sequentially", func(t *testing.T) {
			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))
			output, err := pp.Apply(input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})

		t.Run("concurrently", func(t *testing.T) {
			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))
			output, err := pp.Apply(input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})
	})

	t.Run("can run and fold", func(t *testing.T) {
		input := numbersFromAtoB(0, 100)
		expectedOutput := 92025

		t.Run("sequentially", func(t *testing.T) {
			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))
			output, err := pp.ApplyAndFold(input, p.NewMustFoldOperation(addTwoNumbers))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})

		t.Run("concurrently", func(t *testing.T) {
			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))
			output, err := pp.ApplyAndFold(input, p.NewMustFoldOperation(addTwoNumbers))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})
	})

	t.Run("can stream", func(t *testing.T) {
		input := numbersFromAtoB(0, 100)
		expectedOutput := []int{7, 19, 67, 199, 259, 487, 679, 787, 1027, 1159, 1447, 1939, 2119, 2707, 3139, 3367, 3847, 4099, 4627, 5479, 5779, 6727, 7399, 7747, 8467, 8839, 9607}

		t.Run("sequentially", func(t *testing.T) {
			inCh := make(chan int, 101)
			for _, v := range input {
				inCh <- v
			}
			close(inCh)
			outCh := make(chan int, 101)

			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))

			if err := pp.Stream(inCh, outCh); err != nil {
				t.Fatal(err)
			}
			close(outCh)

			output := make([]int, 0, 100)
			for v := range outCh {
				output = append(output, v)
			}

			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})

		t.Run("concurrently", func(t *testing.T) {
			inCh := make(chan int, 101)
			for _, v := range input {
				inCh <- v
			}
			close(inCh)
			outCh := make(chan int, 101)

			pp := p.New[int]().
				Filter(evenFilter).
				Filter(multiplesOf10Remover).
				MustTransform(squareTransformer).
				MustTransform(adderTransformer(3)).
				Filter(multiplesOfNRemover(3))

			if err := pp.Stream(inCh, outCh); err != nil {
				t.Fatal(err)
			}
			close(outCh)

			output := make([]int, 0, 100)
			for v := range outCh {
				output = append(output, v)
			}
			sort.Ints(output)

			if !reflect.DeepEqual(output, expectedOutput) {
				t.Fatalf("expected: %v, got: %v", expectedOutput, output)
			}
		})
	})
}

func numbersFromAtoB(a, b int) []int {
	result := make([]int, 0, b-a+1)
	for i := a; i <= b; i++ {
		result = append(result, i)
	}
	return result
}

var evenFilter = func(v int) bool {
	return v%2 == 0
}

var multiplesOf10Remover = func(v int) bool {
	if v < 10 {
		return true
	}
	return v%10 != 0
}

var squareTransformer = func(v int) int {
	return v * v
}

var squareButFilter4Transformer = func(v int) (int, bool) {
	if v == 4 {
		return 0, false
	}
	return v * v, true
}

var adderTransformer = func(by int) func(int) int {
	return func(v int) int {
		return v + by
	}
}

var multiplesOfNRemover = func(N int) func(int) bool {
	return func(v int) bool {
		if v < N {
			return true
		}
		return v%N != 0
	}
}

var addTwoNumbers = func(a, b int) int {
	return a + b
}
