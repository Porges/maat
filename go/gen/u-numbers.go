// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package gen

import "math"

type ()

type UintGenerator struct {
	shrinkTarget uint
}

func (g UintGenerator) shrink(value uint, send Shrinkee[uint]) ShrinkResult {
	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := uint(math.Log10(float64(value)))
		if logged != value && logged >= g.shrinkTarget {
			if send(GeneratedValue[uint]{logged, g.shrink}) == Stop {
				return Stopped
			}
		}
	}

	// fast shrink:
	halved := value / 2
	if halved != value && halved >= g.shrinkTarget {
		if send(GeneratedValue[uint]{halved, g.shrink}) == Stop {
			return Stopped
		}
	}

	// slow shrink:
	if value > g.shrinkTarget {
		if send(GeneratedValue[uint]{value - 1, g.shrink}) == Stop {
			return Stopped
		}
	} else if value < g.shrinkTarget {
		if send(GeneratedValue[uint]{value + 1, g.shrink}) == Stop {
			return Stopped
		}
	}

	return Exhausted
}
