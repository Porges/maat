package gen

import (
	"math/rand"
)

type Generator interface {
	Generate(rand *rand.Rand) GeneratedValue
}

type GeneratedValue struct {
	Value    interface{}
	Shrinker Shrinker
}

func (v GeneratedValue) Shrink(s Shrinkee) ShrinkResult {
	return v.Shrinker(v.Value, s)
}

// This is a convoluted way to do generators in go:
// The Shrinker takes a Shrinkee which receives each value. The
// Shrinkee can return Stop to indicate it doesn't want any more values.
// The Shrinker will return Stopped if the Shrinkee returned Stop, or
// otherwise Exhausted.
type Shrinker = func(interface{}, Shrinkee) ShrinkResult
type ShrinkResult int

const (
	Stopped ShrinkResult = iota
	Exhausted
)

type Shrinkee = func(GeneratedValue) ShrinkeeResult
type ShrinkeeResult int

const (
	Stop ShrinkeeResult = iota
	Continue
)
