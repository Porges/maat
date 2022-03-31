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
			return maat.Int(g, "left") == maat.Int(g, "right")
		})

	m.Check("Can use testing.T",
		func(g maat.G, t *testing.T) {
			om := gomega.NewWithT(t)
			om.Expect(maat.Int(g, "x")).To(gomega.Equal(maat.Int(g, "y")))
		})

	m.Boolean("slices",
		func(g maat.G) bool {
			slice := maat.SliceOf[int](g, "values", gen.Int, 2, 100)
			return slice[0] == slice[1]
		})

	m.Boolean("typed slices",
		func(g maat.G) bool {
			slice := maat.SliceOfInt(g, "values", gen.Int, 2, 100)
			return slice[0] == slice[1]
		})
}
