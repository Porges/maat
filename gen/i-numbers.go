// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package gen

import "math"

type ()

type IntGenerator struct {
	shrinkTarget int
}

func (g IntGenerator) shrink(value int, send Shrinkee[int]) ShrinkResult {
	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := int(math.Log10(float64(value)))
		if logged != value {
			if send(GeneratedValue[int]{logged, g.shrink}) == Stop {
				return Stopped
			}
		}
	}

	// fast shrink:
	halved := value / 2
	if halved != value && value > g.shrinkTarget {
		if send(GeneratedValue[int]{halved, g.shrink}) == Stop {
			return Stopped
		}
	}

	// slow shrink:
	if value > g.shrinkTarget {
		if send(GeneratedValue[int]{value - 1, g.shrink}) == Stop {
			return Stopped
		}
	} else if value < g.shrinkTarget {
		if send(GeneratedValue[int]{value + 1, g.shrink}) == Stop {
			return Stopped
		}
	}

	return Exhausted
}
