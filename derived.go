package maat

import (
	"math/rand"

	"github.com/Porges/maat/gen"
	"golang.org/x/exp/slices"
)

func Derive[T any](g *G, name string, deriver func() T) T {
	return RunGenerator[T](g.Runner, name, derived[T]{g, deriver})
}

type derived[T any] struct {
	g       *G
	deriver func() T
}

func (d derived[T]) Generate(r *rand.Rand) gen.GeneratedValue[T] {
	var result T
	recording := recordExecution(r, func(r Runner) {
		oldR := d.g.Runner
		d.g.Runner = r
		defer func() { d.g.Runner = oldR }()
		result = d.deriver()
	})

	return gen.GeneratedValue[T]{
		Value:    result,
		Shrinker: d.shrinkRecording(recording),
	}
}

func (d derived[T]) shrinkRecording(executionRecord recording) gen.Shrinker[T] {
	return func(_ T, send gen.Shrinkee[T]) gen.ShrinkResult {
		for ix := range executionRecord {
			shrinkResult := executionRecord[ix].generated.Shrink(func(gv gen.GeneratedValue[any]) gen.ShrinkeeResult {
				newRecording := replaceRecordedAt[any](executionRecord, ix, gv)
				var result T
				replayExecution(
					newRecording,
					func(r Runner) {
						oldR := d.g.Runner
						d.g.Runner = r
						defer func() { d.g.Runner = oldR }()
						result = d.deriver()
					})

				return send(gen.GeneratedValue[T]{
					Value:    result,
					Shrinker: d.shrinkRecording(newRecording),
				})
			})

			if shrinkResult == gen.Stopped {
				return gen.Stopped
			}
		}

		return gen.Exhausted
	}
}

func replaceRecordedAt[T any](list recording, at int, v gen.GeneratedValue[any]) recording {
	result := slices.Clone(list)
	result[at].generated = v
	return result
}
