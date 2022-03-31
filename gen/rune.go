package gen

import (
	"math/rand"
)

var Rune RuneGenerator = RuneGenerator{}

type RuneGenerator struct{}

var _ Generator[rune] = RuneGenerator{}

func (_ RuneGenerator) Generate(r *rand.Rand) GeneratedValue[rune] {
	return GeneratedValue[rune]{
		Value:    r.Int31n(0x10FFFF + 1),
		Shrinker: Int32Generator{shrinkTarget: 'x'}.shrink,
	}
}
