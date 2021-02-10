package gen

import (
	"math/rand"
)

var Bool BoolGenerator = BoolGenerator{}

type BoolGenerator struct{}

var _ Generator = BoolGenerator{}

func (g BoolGenerator) Generate(r *rand.Rand) GeneratedValue {
	var bytes []byte
	r.Read(bytes)
	return GeneratedValue{
		Value:    bytes[0] & 1,
		Shrinker: shrinkBool,
	}
}

func shrinkBool(v interface{}, send Shrinkee) ShrinkResult {
	value := v.(bool)
	if value {
		if Stop == send(GeneratedValue{
			Value:    false,
			Shrinker: nil,
		}) {
			return Stopped
		}
	}

	return Exhausted
}
