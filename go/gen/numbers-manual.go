package gen

import (
	"math"
	"math/rand"
)

// Manual bits:

func (g IntGenerator) Generate(r *rand.Rand) GeneratedValue[int] {
	return GeneratedValue[int]{r.Int(), g.shrink}
}

var Int IntGenerator = IntGenerator{}

func (g UintGenerator) Generate(r *rand.Rand) GeneratedValue[uint] {
	return GeneratedValue[uint]{uint(r.Uint64()), g.shrink}
}

var Uint UintGenerator = UintGenerator{}

func (g Int64Generator) Generate(r *rand.Rand) GeneratedValue[int64] {
	return GeneratedValue[int64]{int64(r.Uint64()), g.shrink}
}

var Int64 Int64Generator = Int64Generator{}

func (g Uint64Generator) Generate(r *rand.Rand) GeneratedValue[uint64] {
	return GeneratedValue[uint64]{uint64(r.Uint64()), g.shrink}
}

var Uint64 Uint64Generator = Uint64Generator{}

func (g Int32Generator) Generate(r *rand.Rand) GeneratedValue[int32] {
	return GeneratedValue[int32]{int32(r.Uint32()), g.shrink}
}

var Int32 Int32Generator = Int32Generator{}

func (g Uint32Generator) Generate(r *rand.Rand) GeneratedValue[uint32] {
	return GeneratedValue[uint32]{uint32(r.Uint32()), g.shrink}
}

var Uint32 Uint32Generator = Uint32Generator{}

func (g Int16Generator) Generate(r *rand.Rand) GeneratedValue[int16] {
	return GeneratedValue[int16]{int16(r.Intn(math.MaxInt16 + 1)), g.shrink}
}

var Int16 Int16Generator = Int16Generator{}

func (g Uint16Generator) Generate(r *rand.Rand) GeneratedValue[uint16] {
	return GeneratedValue[uint16]{uint16(r.Intn(math.MaxUint16 + 1)), g.shrink}
}

var Uint16 Uint16Generator = Uint16Generator{}

func (g Int8Generator) Generate(r *rand.Rand) GeneratedValue[int8] {
	return GeneratedValue[int8]{int8(r.Intn(math.MaxInt8 + 1)), g.shrink}
}

var Int8 Int8Generator = Int8Generator{}

func (g Uint8Generator) Generate(r *rand.Rand) GeneratedValue[uint8] {
	return GeneratedValue[uint8]{uint8(r.Intn(math.MaxUint8 + 1)), g.shrink}
}

var (
	Uint8 Uint8Generator = Uint8Generator{}
	Byte  Uint8Generator = Uint8Generator{}
)
