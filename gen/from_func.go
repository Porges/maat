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
