package maat

import (
	"fmt"
	"math/rand"

	"github.com/Porges/maat/gen"
)

type Runner struct {
	// this should really be 3 structs that implement a common interface
	// but you can't have a generic method in go
	mode runnerMode

	recordIx int
	rand     *rand.Rand

	recording *recording
}

type runnerMode int

const (
	test runnerMode = iota
	record
	replay
)

func Run[T any](r *Runner, name string, generator gen.Generator[T]) T {
	// see note on `Runner` struct
	switch r.mode {
	case test:
		panic("TODO")
	case record:
		generated := generator.Generate(r.rand)
		record := &generationRecord[T]{
			name:      name,
			generator: generator,
			generated: generated,
		}

		*r.recording = append(*r.recording, record)
		return generated.Value

	case replay:
		record := (*r.recording)[r.recordIx].(*generationRecord[T])
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
	default:
		panic("Runner mode undefined")
	}
}

type recording []generationRecordInterface

type generationRecord[T any] struct {
	name      string
	generator gen.Generator[T]
	generated gen.GeneratedValue[T]
	depth     int
}

func (me *generationRecord[T]) AttemptShrink(runShrunk func() bool) bool {
	generated := me.generated

	if generated.Shrinker == nil {
		return false
	}

	shrinkResult := generated.Shrink(func(possible gen.GeneratedValue[T]) gen.ShrinkeeResult {
		me.generated = possible
		success := runShrunk()
		if success {
			// restore old value
			me.generated = generated
			return gen.Continue
		} else {
			return gen.Stop
		}
	})

	return shrinkResult == gen.Stopped
}

func (me *generationRecord[T]) Name() string {
	return me.name
}

func (me *generationRecord[T]) Value() any {
	return me.generated.Value
}

func (me *generationRecord[T]) Clone() generationRecordInterface {
	clone := *me
	return &clone
}

var _ generationRecordInterface = &generationRecord[int]{}

type generationRecordInterface interface {
	AttemptShrink(runShrunk func() bool) bool // result = did shrink
	Name() string
	Value() any
	Clone() generationRecordInterface
}

func recordExecution(randSource rand.Source, toRecord func(*Runner)) recording {
	var result recording
	toRecord(&Runner{mode: record, rand: rand.New(randSource), recording: &result})
	return result
}

func replayExecution(recording recording, toReplay func(*Runner)) {
	toReplay(&Runner{mode: replay, recording: &recording})
}
