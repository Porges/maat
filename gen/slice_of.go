package gen

import (
	"math/rand"
)

func SliceOf(element Generator, minSize int, maxSize int) Generator {
	if maxSize < minSize {
		panic("maxSize must be greater than or equal to minSize")
	}

	return sliceOf{element, minSize, maxSize}
}

type sliceOf struct {
	inner   Generator
	minSize int
	maxSize int
}

func (s sliceOf) Generate(r *rand.Rand) GeneratedValue {
	size := s.minSize
	if s.maxSize > s.minSize {
		size = s.minSize + r.Intn(s.maxSize-s.minSize)
	}

	result := make([]interface{}, size)
	shrinkers := make([]Shrinker, size)

	for ix := range result {
		generated := s.inner.Generate(r)
		result[ix] = generated.Value
		shrinkers[ix] = generated.Shrinker
	}

	return GeneratedValue{
		Value:    result,
		Shrinker: sliceShrinker(s.minSize, shrinkers),
	}
}

func sliceShrinker(minSize int, shrinkers []Shrinker) Shrinker {
	return func(v interface{}, send Shrinkee) ShrinkResult {
		value := v.([]interface{})
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
				if Stop == send(GeneratedValue{
					Value:    removeValueAt(value, ix),
					Shrinker: sliceShrinker(minSize, removeShrinkerAt(shrinkers, ix)),
				}) {
					return Stopped
				}
			}

			// or shrink the value
			if Stopped == shrinkers[ix](value[ix], func(shrunk GeneratedValue) ShrinkeeResult {
				cloned := make([]interface{}, len(value))
				for innerIx := range cloned {
					if innerIx == ix {
						cloned[innerIx] = shrunk.Value
					} else {
						cloned[innerIx] = value[innerIx]
					}
				}

				return send(GeneratedValue{
					Value:    cloned,
					Shrinker: sliceShrinker(minSize, replaceShrinkerAt(shrinkers, ix, shrunk.Shrinker)),
				})
			}) {
				return Stopped
			}
		}

		return Exhausted
	}
}

func replaceShrinkerAt(shrinkers []Shrinker, at int, with Shrinker) []Shrinker {
	result := make([]Shrinker, len(shrinkers))
	copy(result, shrinkers)
	result[at] = with
	return result
}

func removeValueAt(shrinkers []interface{}, at int) []interface{} {
	result := make([]interface{}, len(shrinkers)-1)
	copy(result[:at], shrinkers[:at])
	copy(result[at:], shrinkers[at+1:])
	return result
}

func removeShrinkerAt(shrinkers []Shrinker, at int) []Shrinker {
	result := make([]Shrinker, len(shrinkers)-1)
	copy(result[:at], shrinkers[:at])
	copy(result[at:], shrinkers[at+1:])
	return result
}
