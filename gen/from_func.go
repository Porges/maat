package gen

import (
	"math/rand"
)

type FuncGen struct {
	generate func(r *rand.Rand) interface{}
}

var _ Generator = FuncGen{}

func FromFunc(generate func(r *rand.Rand) interface{}) FuncGen {
	return FuncGen{generate}
}

func (fg FuncGen) Generate(r *rand.Rand) GeneratedValue {
	return GeneratedValue{fg.generate(r), nil}
}
