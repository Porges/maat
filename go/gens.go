package maat

import (
	"github.com/Porges/maat/go/gen"
)

type G struct {
	runner *Runner
}

func Byte(g G, name string) byte {
	return Run[byte](g.runner, name, gen.Byte)
}

func Bool(g G, name string) bool {
	return Run[bool](g.runner, name, gen.Bool)
}

func Int(g G, name string) int {
	return Run[int](g.runner, name, gen.Int)
}

func SliceOf[T any](g G, name string, element gen.Generator[T], minSize int, maxSize int) []T {
	return Run(g.runner, name, gen.SliceOf(element, minSize, maxSize))
}

func SliceOfInt(g G, name string, element gen.IntGenerator, minSize int, maxSize int) []int {
	return Derive(
		g,
		name,
		func() []int {
			return SliceOf[int](g, "values", element, minSize, maxSize)
		})
}

func String(g G, name string, minSize int, maxSize int) string {
	return Derive(
		g,
		name,
		func() string {
			runes := SliceOf[rune](g, "chars", gen.Rune, 1, 10)
			return string(runes)
		})
}
