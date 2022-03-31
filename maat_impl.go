package maat

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"github.com/Porges/maat/gen"
)

type impl struct {
	t      *testing.T
	config config
}

func (i impl) Check(name string, checkable func(g G, t *testing.T)) bool {
	return i.t.Run(name+i.nameSuffix(), func(t *testing.T) {
		source := rand.NewSource(0)

		var executionRecord recording
		defer func() {
			if t.Failed() && executionRecord != nil {
				t.Logf("%s", i.shrink(name, executionRecord, checkable))
			}
		}()

		for iteration := 0; iteration < int(i.config.Iterations); iteration++ {
			newSeed := time.Now().Unix() // TODO: source stored seeds from somewhere
			source.Seed(newSeed)
			executionRecord = nil // reset here so if next line panics/exits it is nil
			executionRecord = recordExecution(rand.New(source), func(r Runner) { checkable(G{r}, t) })
			if t.Failed() {
				// if user called Fail instead of FailNow, we want to stop
				t.FailNow()
				break
			}
		}
	})
}

func (i impl) Boolean(name string, checkable func(g G) bool) bool {
	return i.Check(name, func(g G, t *testing.T) {
		if !checkable(g) {
			t.Fail()
		}
	})
}

func (i impl) shrink(name string, executionRecord recording, checkable func(g G, t *testing.T)) string {
	initial := slices.Clone(executionRecord)

	shrinks := 0
	for {
		shrank := false
		for ix := range executionRecord {
			generated := executionRecord[ix].generated
			if generated.Shrinker != nil {
				shrinkResult := generated.Shrink(func(possible gen.GeneratedValue[any]) gen.ShrinkeeResult {
					shrinks += 1
					executionRecord[ix].generated = possible
					success := i.t.Run(fmt.Sprintf("%s (shrink attempt %v)", name, shrinks), func(t *testing.T) {
						replayExecution(executionRecord, func(r Runner) { checkable(G{r}, t) })
					})

					if success {
						// restore old value
						executionRecord[ix].generated = generated
						return gen.Continue
					} else {
						return gen.Stop
					}
				})

				if shrinkResult == gen.Stopped {
					shrank = true
				}
			}
		}

		if !shrank {
			break
		}
	}

	var original []string
	for _, r := range initial {
		original = append(original, fmt.Sprintf("%s: %s", r.name, i.config.Printer(r.generated.Value)))
	}

	var report []string
	for _, r := range executionRecord {
		report = append(report, fmt.Sprintf("%s: %s", r.name, i.config.Printer(r.generated.Value)))
	}

	return fmt.Sprintf("\n[Maat] Shrunk failure:\n%s\n\n[Maat] Original failure:\n%s\n\n", strings.Join(report, "\n"), strings.Join(original, "\n"))
}

type usageError struct {
	msg string
}

var _ error = usageError{}

func (ue usageError) Error() string {
	return ue.msg
}

func (i impl) nameSuffix() string {
	return fmt.Sprintf(" (%v random tests)", i.config.Iterations)
}
