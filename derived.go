package maat

import (
	"math/rand"

	"github.com/Porges/maat/gen"
)

func (g *G) Derived(name string, deriver func() interface{}) interface{} {
	return g.RunGenerator(name, derived{g, deriver})
}

type derived struct {
	g       *G
	deriver func() interface{}
}

func (d derived) Generate(r *rand.Rand) gen.GeneratedValue {

	var result interface{}
	recording := recordExecution(r, func(r Runner) {
		oldR := d.g.Runner
		d.g.Runner = r
		defer func() { d.g.Runner = oldR }()
		result = d.deriver()
	})

	return gen.GeneratedValue{
		Value:    result,
		Shrinker: d.shrinkRecording(recording),
	}
}

func (d derived) shrinkRecording(executionRecord recording) gen.Shrinker {
	return func(_ interface{}, send gen.Shrinkee) gen.ShrinkResult {
		for ix := range executionRecord {
			shrinkResult := executionRecord[ix].generated.Shrink(func(gv gen.GeneratedValue) gen.ShrinkeeResult {
				newRecording := replaceRecordedAt(executionRecord, ix, gv)
				var result interface{}
				replayExecution(
					newRecording,
					func(r Runner) {
						oldR := d.g.Runner
						d.g.Runner = r
						defer func() { d.g.Runner = oldR }()
						result = d.deriver()
					})

				return send(gen.GeneratedValue{
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

func replaceRecordedAt(list recording, at int, v gen.GeneratedValue) recording {
	result := make(recording, len(list))
	copy(result, list)
	result[at].generated = v
	return result
}
