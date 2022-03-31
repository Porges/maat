package gen

import (
	"math/rand"
)

func SliceOf[T any](element Generator[T], minSize int, maxSize int) Generator[[]T] {
	if maxSize < minSize {
		panic("maxSize must be greater than or equal to minSize")
	}

	return sliceOf[T]{element, minSize, maxSize}
}

type sliceOf[T any] struct {
	inner   Generator[T]
	minSize int
	maxSize int
}

func (s sliceOf[T]) Generate(r *rand.Rand) GeneratedValue[[]T] {
	size := s.minSize
	if s.maxSize > s.minSize {
		size = s.minSize + r.Intn(s.maxSize-s.minSize)
	}

	result := make([]T, size)
	shrinkers := make([]Shrinker[T], size)

	for ix := range result {
		generated := s.inner.Generate(r)
		result[ix] = generated.Value
		shrinkers[ix] = generated.Shrinker
	}

	return GeneratedValue[[]T]{
		Value:    result,
		Shrinker: sliceShrinker(s.minSize, shrinkers),
	}
}

func sliceShrinker[T any](minSize int, shrinkers []Shrinker[T]) Shrinker[[]T] {
	return func(value []T, send Shrinkee[[]T]) ShrinkResult {
		for ix := range value {
			// take everything up to this point
			/* TODO: make it work?
			if ix > 0 && ix >= minSize {
				sliced := value[:ix]
				if !send(GeneratedValue{
					Value:  sliced,
					Shrink: sliceShrinker(sliced, minSize, shrinkers[:ix]),
				}) {
					return false
				}
			}
			*/

			// drop the value entirely
			if len(value) > minSize {
				if Stop == send(GeneratedValue[[]T]{
					Value:    removeValueAt(value, ix),
					Shrinker: sliceShrinker(minSize, removeValueAt(shrinkers, ix)),
				}) {
					return Stopped
				}
			}

			// or shrink the value
			if Stopped == shrinkers[ix](value[ix], func(shrunk GeneratedValue[T]) ShrinkeeResult {
				cloned := make([]T, len(value))
				for innerIx := range cloned {
					if innerIx == ix {
						cloned[innerIx] = shrunk.Value
					} else {
						cloned[innerIx] = value[innerIx]
					}
				}

				return send(GeneratedValue[[]T]{
					Value:    cloned,
					Shrinker: sliceShrinker(minSize, replaceValueAt(shrinkers, ix, shrunk.Shrinker)),
				})
			}) {
				return Stopped
			}
		}

		return Exhausted
	}
}

func replaceValueAt[T any](xs []T, at int, with T) []T {
	result := make([]T, len(xs))
	copy(result, xs)
	result[at] = with
	return result
}

func removeValueAt[T any](xs []T, at int) []T {
	result := make([]T, len(xs)-1)
	copy(result[:at], xs[:at])
	copy(result[at:], xs[at+1:])
	return result
}
