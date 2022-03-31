// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package gen

import "math"

type ()

type Int8Generator struct {
	shrinkTarget int8
}

func (g Int8Generator) shrink(value int8, send Shrinkee[int8]) ShrinkResult {
	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := int8(math.Log10(float64(value)))
		if logged != value && logged >= g.shrinkTarget {
			if send(GeneratedValue[int8]{logged, g.shrink}) == Stop {
				return Stopped
			}
		}
	}

	// fast shrink:
	halved := value / 2
	if halved != value && halved >= g.shrinkTarget {
		if send(GeneratedValue[int8]{halved, g.shrink}) == Stop {
			return Stopped
		}
	}

	// slow shrink:
	if value > g.shrinkTarget {
		if send(GeneratedValue[int8]{value - 1, g.shrink}) == Stop {
			return Stopped
		}
	} else if value < g.shrinkTarget {
		if send(GeneratedValue[int8]{value + 1, g.shrink}) == Stop {
			return Stopped
		}
	}

	return Exhausted
}
