package test

import (
	"testing"

	"github.com/Porges/maat"
	"github.com/Porges/maat/gen"
	"github.com/kr/pretty"
	"github.com/onsi/gomega"
)

func prettyPrint(value interface{}) string {
	return pretty.Sprint(value)
}

func TestMaat(t *testing.T) {
	m := maat.New(t, maat.PrettyPrinter(prettyPrint))

	m.Boolean("Can use simple assertions",
		func(g maat.G) bool {
			return g.Int("left") == g.Int("right")
		})

	m.Check("Can use testing.T",
		func(g maat.G, t *testing.T) {
			om := gomega.NewWithT(t)
			om.Expect(g.Int("x")).To(gomega.Equal(g.Int("y")))
		})

	m.Boolean("slices",
		func(g maat.G) bool {
			slice := g.SliceOf("values", gen.Int, 2, 100)
			return slice[0].(int) == slice[1].(int)
		})

	m.Boolean("typed slices",
		func(g maat.G) bool {
			slice := g.SliceOfInt("values", gen.Int, 2, 100)
			return slice[0] == slice[1]
		})
}
