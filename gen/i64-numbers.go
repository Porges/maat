// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package gen

import "math"

type Int64Generator struct {
	shrinkTarget int64
}

func (g Int64Generator) shrink(v interface{}, send Shrinkee) ShrinkResult {
	value := v.(int64)

	// very fast shrink:
	if value > 0 && value > g.shrinkTarget {
		logged := int64(math.Log10(float64(value)))
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
