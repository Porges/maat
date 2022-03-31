package maat

import (
	"fmt"
	"math/rand"

	"github.com/Porges/maat/gen"
)

type Runner interface {
	RunGenerator(name string, generator gen.Generator[any]) any
}

func RunGenerator[T any](r Runner, name string, generator gen.Generator[T]) T {
	return r.RunGenerator(name, gen.ToAny(generator)).(T)
}

type recording []generationRecord

type generationRecord struct {
	name      string
	generator gen.Generator[any]
	generated gen.GeneratedValue[any]
	depth     int
}

func recordExecution(randSource rand.Source, toRecord func(Runner)) recording {
	var result recording
	r := &recordRunner{rand: rand.New(randSource), recordTo: &result}
	toRecord(r)
	return result
}

type recordRunner struct {
	rand     *rand.Rand
	recordTo *recording
}

func (r *recordRunner) RunGenerator(name string, generator gen.Generator[any]) any {
	generated := generator.Generate(r.rand)
	record := generationRecord{
		name:      name,
		generator: generator,
		generated: generated,
	}

	*r.recordTo = append(*r.recordTo, record)
	return generated.Value
}

func replayExecution(executionRecord recording, toReplay func(Runner)) {
	r := &replayRunner{executionRecord: executionRecord}
	toReplay(r)
}

type replayRunner struct {
	recordIx        int
	executionRecord recording
}

func (r *replayRunner) RunGenerator(name string, generator gen.Generator[any]) any {
	record := r.executionRecord[r.recordIx]
	r.recordIx++
	// TODO: need Equals
	// if record.generator != generator {
	//panic(usageError{"Different generator type"})
	//}

	if record.name != name {
		panic(usageError{fmt.Sprintf("Different generator name, was %s, is %s.", record.name, name)})
	} else {
		return record.generated.Value
	}
}
