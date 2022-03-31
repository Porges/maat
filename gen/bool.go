package gen

import (
	"math/rand"
)

var Bool BoolGenerator = BoolGenerator{}

type BoolGenerator struct{}

var _ Generator[bool] = BoolGenerator{}

func (g BoolGenerator) Generate(r *rand.Rand) GeneratedValue[bool] {
	var bytes []byte
	r.Read(bytes)
	return GeneratedValue[bool]{
		Value:    (bytes[0] & 1) != 0,
		Shrinker: shrinkBool,
	}
}

func shrinkBool(value bool, send Shrinkee[bool]) ShrinkResult {
	if value {
		if Stop == send(GeneratedValue[bool]{
			Value:    false,
			Shrinker: nil,
		}) {
			return Stopped
		}
	}

	return Exhausted
}
