package maat

import "github.com/Porges/maat/gen"

type G struct {
	Runner
}

func (g G) Byte(name string) byte {
	return g.RunGenerator(name, gen.Byte).(byte)
}

func (g G) Bool(name string) bool {
	return g.RunGenerator(name, gen.Bool).(bool)
}

func (g G) Int(name string) int {
	return g.RunGenerator(name, gen.Int).(int)
}

func (g G) SliceOf(name string, element gen.Generator, minSize int, maxSize int) []interface{} {
	return g.RunGenerator(name, gen.SliceOf(element, minSize, maxSize)).([]interface{})
}

func (g G) SliceOfInt(name string, element gen.IntGenerator, minSize int, maxSize int) []int {
	return g.Derived(
		name,
		func() interface{} {
			inner := g.SliceOf("values", element, minSize, maxSize)
			result := make([]int, len(inner))
			for ix := range inner {
				result[ix] = inner[ix].(int)
			}

			return result
		}).([]int)
}

func (g G) String(name string, minSize int, maxSize int) string {
	return g.Derived(
		name,
		func() interface{} {
			chars := g.SliceOf("chars", gen.Char, 1, 10)
			runes := make([]rune, len(chars))
			for ix := range chars {
				runes[ix] = chars[ix].(rune)
			}
			return string(runes)
		}).(string)
}
