// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package gen

import "math"

type ()

type Int16Generator struct {
	shrinkTarget int16
}

func (g Int16Generator) shrink(value int16, send Shrinkee[int16]) ShrinkResult {
	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := int16(math.Log10(float64(value)))
		if logged != value && logged >= g.shrinkTarget {
			if send(GeneratedValue[int16]{logged, g.shrink}) == Stop {
				return Stopped
			}
		}
	}

	// fast shrink:
	halved := value / 2
	if halved != value && halved >= g.shrinkTarget {
		if send(GeneratedValue[int16]{halved, g.shrink}) == Stop {
			return Stopped
		}
	}

	// slow shrink:
	if value > g.shrinkTarget {
		if send(GeneratedValue[int16]{value - 1, g.shrink}) == Stop {
			return Stopped
		}
	} else if value < g.shrinkTarget {
		if send(GeneratedValue[int16]{value + 1, g.shrink}) == Stop {
			return Stopped
		}
	}

	return Exhausted
}
