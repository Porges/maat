package test

import (
	"testing"

	"github.com/Porges/maat/go"
)

// You can define your generators easily:
type ColorType struct{ r, g, b byte }

func Color(g maat.G, name string) ColorType {
	return maat.Derive(
		g,
		name,
		func() ColorType {
			return ColorType{maat.Byte(g, "r"), maat.Byte(g, "g"), maat.Byte(g, "b")}
		})
}

// ... and generators can depend on other custom generators:
type DogType struct {
	name  string
	color ColorType
}

func Dog(g maat.G, name string) DogType {
	return maat.Derive(
		g,
		name,
		func() DogType {
			return DogType{
				name:  maat.String(g, "name", 1, 10),
				color: Color(g, "color"),
			}
		})
}

// Example test, showing that shrinking automatically works for custom types:
func TestCustom(t *testing.T) {
	m := maat.New(t, maat.PrettyPrinter(prettyPrint))
	m.Boolean(
		"dog generator",
		func(g maat.G) bool {
			return Dog(g, "d1").name == Dog(g, "d2").name
		})
}
