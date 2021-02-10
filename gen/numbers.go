package gen

import (
	"math"

	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=$GOFILE -out=i-$GOFILE -pkg gen gen "_T=Int _t=int"
//go:generate genny -in=$GOFILE -out=u-$GOFILE -pkg gen gen "_T=Uint _t=uint"

//go:generate genny -in=$GOFILE -out=i8-$GOFILE -pkg gen gen "_T=Int8 _t=int8"
//go:generate genny -in=$GOFILE -out=u8-$GOFILE -pkg gen gen "_T=Uint8 _t=uint8"

//go:generate genny -in=$GOFILE -out=i16-$GOFILE -pkg gen gen "_T=Int16 _t=int16"
//go:generate genny -in=$GOFILE -out=u16-$GOFILE -pkg gen gen "_T=Uint16 _t=uint16"

//go:generate genny -in=$GOFILE -out=i32-$GOFILE -pkg gen gen "_T=Int32 _t=int32"
//go:generate genny -in=$GOFILE -out=u32-$GOFILE -pkg gen gen "_T=Uint32 _t=uint32"

//go:generate genny -in=$GOFILE -out=i64-$GOFILE -pkg gen gen "_T=Int64 _t=int64"
//go:generate genny -in=$GOFILE -out=u64-$GOFILE -pkg gen gen "_T=Uint64 _t=uint64"

type _T generic.Type
type _t generic.Number

type _TGenerator struct {
	shrinkTarget _t
}

func (g _TGenerator) shrink(v interface{}, send Shrinkee) ShrinkResult {
	value := v.(_t)

	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := _t(math.Log10(float64(value)))
		if logged != value {
			if send(GeneratedValue{logged, g.shrink}) == Stop {
				return Stopped
			}
		}
	}

	// fast shrink:
	halved := value / 2
	if halved != value && value > g.shrinkTarget {
		if send(GeneratedValue{halved, g.shrink}) == Stop {
			return Stopped
		}
	}

	// slow shrink:
	if value > g.shrinkTarget {
		if send(GeneratedValue{value - 1, g.shrink}) == Stop {
			return Stopped
		}
	} else if value < g.shrinkTarget {
		if send(GeneratedValue{value + 1, g.shrink}) == Stop {
			return Stopped
		}
	}

	return Exhausted
}
