package gen

import (
	"math/rand"
)

func OneOf(generators ...Generator) Generator {
	return oneOf{generators}
}

type oneOf struct {
	inner []Generator
}

func (o oneOf) Generate(r *rand.Rand) GeneratedValue {
	ix := r.Intn(len(o.inner))
	return o.inner[ix].Generate(r)
}
