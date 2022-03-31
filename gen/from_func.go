package gen

import (
	"math/rand"
)

type FuncGen[T any] struct {
	generate func(r *rand.Rand) T
}

var _ Generator[any] = FuncGen[any]{}

func FromFunc[T any](generate func(r *rand.Rand) T) FuncGen[T] {
	return FuncGen[T]{generate}
}

func (fg FuncGen[T]) Generate(r *rand.Rand) GeneratedValue[T] {
	return GeneratedValue[T]{fg.generate(r), nil}
}

func GRToAny[T any](gr GeneratedValue[T]) GeneratedValue[any] {
	return GeneratedValue[any]{
		Value: gr.Value,
		Shrinker: func(value any, shrinkee Shrinkee[any]) ShrinkResult {
			return gr.Shrinker(value.(T), func(x GeneratedValue[T]) ShrinkeeResult {
				return shrinkee(GRToAny(x))
			})
		},
	}
}

func ToAny[T any](g Generator[T]) Generator[any] {
	return anyGen{
		generate: func(r *rand.Rand) GeneratedValue[any] {
			return GRToAny(g.Generate(r))
		},
	}
}

type anyGen struct {
	generate func(r *rand.Rand) GeneratedValue[any]
}

func (a anyGen) Generate(r *rand.Rand) GeneratedValue[any] {
	return a.generate(r)
}

var _ Generator[any] = anyGen{}
