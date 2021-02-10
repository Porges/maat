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

// ... and then wrap the functions:
func (m MyMaat) Boolean(name string, checkable func(g G) bool) bool {
	return m.Maat.Boolean(name, func(g maat.G) bool { return checkable(G{g}) })
}

// Now you can define your generators easily:
type Color struct{ r, g, b byte }

func (g G) Color(name string) Color {
	return g.Derived(
		name,
		func() interface{} {
			return Color{g.Byte("r"), g.Byte("g"), g.Byte("b")}
		}).(Color)
}

// ... and generators can depend on other custom generators:
type Dog struct {
	name  string
	color Color
}

func (g G) Dog(name string) Dog {
	return g.Derived(
		name,
		func() interface{} {
			return Dog{
				name:  g.String("name", 1, 10),
				color: g.Color("color"),
			}
		}).(Dog)
}

// Example test:
func TestCustom(t *testing.T) {
	m := NewMaat(t)
	m.Boolean(
		"dog generator",
		func(g G) bool {
			return g.Dog("d1").name == g.Dog("d2").name
		})
}
