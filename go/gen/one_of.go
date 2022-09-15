package gen

import (
	"math/rand"
)

func OneOf[T any](generators ...Generator[T]) Generator[T] {
	return oneOf[T]{generators}
}

type oneOf[T any] struct {
	inner []Generator[T]
}

func (o oneOf[T]) Generate(r *rand.Rand) GeneratedValue[T] {
	ix := r.Intn(len(o.inner))
	return o.inner[ix].Generate(r)
}
