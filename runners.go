package maat

import (
	"math/rand"

	"github.com/Porges/maat/gen"
)

func recordExecution(randSource rand.Source, toRecord func(Runner)) recording {
	var result recording
	r := &recordRunner{rand: rand.New(randSource), recordTo: &result}
	toRecord(r)
	return result
}

func replayExecution(executionRecord recording, toReplay func(Runner)) {
	r := &replayRunner{executionRecord: executionRecord}
	toReplay(r)
}

type Runner interface {
	RunGenerator(name string, generator gen.Generator) interface{}
}

type recording []generationRecord

type generationRecord struct {
	name      string
	generator gen.Generator
	generated gen.GeneratedValue
	depth     int
}

type recordRunner struct {
	rand     *rand.Rand
	recordTo *recording
}

func (r *recordRunner) RunGenerator(name string, generator gen.Generator) interface{} {
	generated := generator.Generate(r.rand)
	record := generationRecord{
		name:      name,
		generator: generator,
		generated: generated,
	}

	*r.recordTo = append(*r.recordTo, record)
	return generated.Value
}

type replayRunner struct {
	recordIx        int
	executionRecord recording
}

func (r *replayRunner) RunGenerator(name string, generator gen.Generator) interface{} {
	record := r.executionRecord[r.recordIx]
	r.recordIx++
	// TODO: need Equals
	// if record.generator != generator {
	//panic(usageError{"Different generator type"})
	//}

	if record.name != name {
		panic(usageError{"Different generator name"})
	} else {
		return record.generated.Value
	}
}
