package maat

import (
	"math/rand"

	"github.com/Porges/maat/go/gen"
	"golang.org/x/exp/slices"
)

func Derive[T any](g G, name string, deriver func() T) T {
	return Run[T](g.runner, name, derived[T]{g, name, deriver})
}

type derived[T any] struct {
	g       G
	name    string
	deriver func() T
}

func (d derived[T]) Generate(r *rand.Rand) gen.GeneratedValue[T] {
	var result T
	recording := recordExecution(r, func(r *Runner) {
		oldR := *d.g.runner
		*d.g.runner = *r
		defer func() { *d.g.runner = oldR }()
		result = d.deriver()
	})

	return gen.GeneratedValue[T]{
		Value:    result,
		Shrinker: d.shrinkRecording(recording),
	}
}

func (d derived[T]) shrinkRecording(executionRecord recording) gen.Shrinker[T] {
	return func(_ T, send gen.Shrinkee[T]) gen.ShrinkResult {
		for _, generationRecord := range executionRecord {
			stopped := generationRecord.AttemptShrink(func() bool {
				var result T
				replayExecution(
					executionRecord,
					func(r *Runner) {
						oldR := *d.g.runner
						*d.g.runner = *r
						defer func() { *d.g.runner = oldR }()
						result = d.deriver()
					})

				return send(gen.GeneratedValue[T]{
					Value:    result,
					Shrinker: d.shrinkRecording(deepClone(executionRecord)),
				}) == gen.Continue
			})

			if stopped {
				return gen.Stopped
			}
		}

		return gen.Exhausted
	}
}

func deepClone(list recording) recording {
	result := slices.Clone(list)
	for ix := range result {
		result[ix] = result[ix].Clone()
	}

	return result
}
