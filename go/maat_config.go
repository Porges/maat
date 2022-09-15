package maat

import "fmt"

type config struct {
	Iterations uint16
	Printer    func(value interface{}) string
}

var defaultConfig config = config{
	Iterations: 100,
	Printer: func(value interface{}) string {
		return fmt.Sprintf("%v", value)
	},
}

type Option func(c *config)

func Iterations(count uint16) Option {
	return func(c *config) {
		c.Iterations = count
	}
}

func PrettyPrinter(printer func(value interface{}) string) Option {
	return func(c *config) {
		c.Printer = printer
	}
}
