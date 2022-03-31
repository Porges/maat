package maat

import (
	"github.com/Porges/maat/gen"
)

type G struct {
	Runner
}

func Byte(g G, name string) byte {
	return RunGenerator[byte](g.Runner, name, gen.Byte)
}

func Bool(g G, name string) bool {
	return RunGenerator[bool](g.Runner, name, gen.Bool)
}

func Int(g G, name string) int {
	return RunGenerator[int](g.Runner, name, gen.Int)
}

func SliceOf[T any](g *G, name string, element gen.Generator[T], minSize int, maxSize int) []T {
	return RunGenerator(g.Runner, name, gen.SliceOf(element, minSize, maxSize))
}

func SliceOfInt(g *G, name string, element gen.IntGenerator, minSize int, maxSize int) []int {
	return Derive(
		g,
		name,
		func() []int {
			inner := SliceOf(g, "values", gen.ToAny[int](element), minSize, maxSize)
			result := make([]int, len(inner))
			for ix := range inner {
				result[ix] = inner[ix].(int)
			}

			return result
		})
}

func String(g *G, name string, minSize int, maxSize int) string {
	return Derive(
		g,
		name,
		func() string {
			chars := SliceOf(g, "chars", gen.ToAny[rune](gen.Rune), 1, 10)
			runes := make([]rune, len(chars))
			for ix := range chars {
				runes[ix] = chars[ix].(rune)
			}
			return string(runes)
		})
}
