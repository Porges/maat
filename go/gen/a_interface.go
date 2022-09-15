package gen

import (
	"math/rand"
)

type Generator[T any] interface {
	Generate(rand *rand.Rand) GeneratedValue[T]
}

type GeneratedValue[T any] struct {
	Value    T
	Shrinker Shrinker[T]
}

func (v GeneratedValue[T]) Shrink(s Shrinkee[T]) ShrinkResult {
	return v.Shrinker(v.Value, s)
}

// This is a convoluted way to do generators in go:
// The Shrinker takes a Shrinkee which receives each value. The
// Shrinkee can return Stop to indicate it doesn't want any more values.
// The Shrinker will return Stopped if the Shrinkee returned Stop, or
// otherwise Exhausted.
type Shrinker[T any] func(T, Shrinkee[T]) ShrinkResult
type ShrinkResult int

const (
	Stopped ShrinkResult = iota
	Exhausted
)

type (
	Shrinkee[T any] func(GeneratedValue[T]) ShrinkeeResult
	ShrinkeeResult  int
)

const (
	Stop ShrinkeeResult = iota
	Continue
)
