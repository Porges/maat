package test

import (
	"testing"

	"github.com/Porges/maat"
)

// If you have a lot of custom types to generate, define a wrapper around Maat:
type MyMaat struct{ maat.Maat }
type G struct{ maat.G }

// ... and a way to build it:
func NewMaat(t *testing.T) MyMaat {
	return MyMaat{maat.New(t, maat.PrettyPrinter(prettyPrint))}
}

// ... and then wrap the two functions:
func (m MyMaat) Boolean(name string, checkable func(g G) bool) bool {
	return m.Maat.Boolean(name, func(g maat.G) bool { return checkable(G{g}) })
}

func (m MyMaat) Check(name string, checkable func(g G)) bool {
	return m.Maat.Check(name, func(g maat.G, t *testing.T) { checkable(G{g}) })
}

// Now you can define your generators easily:
type ColorType struct{ r, g, b byte }

func Color(g *G, name string) ColorType {
	return maat.Derive(
		&g.G,
		name,
		func() ColorType {
			return ColorType{maat.Byte(g.G, "r"), maat.Byte(g.G, "g"), maat.Byte(g.G, "b")}
		})
}

// ... and generators can depend on other custom generators:
type DogType struct {
	name  string
	color ColorType
}

func Dog(g *G, name string) DogType {
	return maat.Derive(
		&g.G,
		name,
		func() DogType {
			return DogType{
				name:  maat.String(&g.G, "name", 1, 10),
				color: Color(g, "color"),
			}
		})
}

// Example test:
func TestCustom(t *testing.T) {
	m := NewMaat(t)
	m.Boolean(
		"dog generator",
		func(g G) bool {
			return Dog(&g, "d1").name == Dog(&g, "d2").name
		})
}
