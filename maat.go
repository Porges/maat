package maat

import (
	"testing"
)

func New(t *testing.T, options ...Option) Maat {
	cfg := defaultConfig
	for _, option := range options {
		option(&cfg)
	}

	return impl{t, cfg}
}

type Maat interface {
	Check(name string, checkable func(g G, t *testing.T)) bool
	Boolean(name string, checkable func(g G) bool) bool
}
