package gen

import (
	"math/rand"
)

var Char CharGenerator = CharGenerator{}

type CharGenerator struct{}

var _ Generator = CharGenerator{}

func (_ CharGenerator) Generate(r *rand.Rand) GeneratedValue {
	return GeneratedValue{
		Value:    r.Int31n(0x10FFFF + 1),
		Shrinker: Int32Generator{shrinkTarget: 'x'}.shrink,
	}
}
