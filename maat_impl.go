package maat

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type impl struct {
	t      *testing.T
	config config
}

func (i impl) Check(name string, checkable func(g G, t *testing.T)) bool {
	return i.t.Run(name+i.nameSuffix(), func(t *testing.T) {
		source := rand.NewSource(0)

		var recording recording
		defer func() {
			if t.Failed() && recording != nil {
				t.Logf("%s", i.shrink(name, recording, checkable))
			}
		}()

		for iteration := 0; iteration < int(i.config.Iterations); iteration++ {
			newSeed := time.Now().Unix() // TODO: source stored seeds from somewhere
			source.Seed(newSeed)
			recording = nil // reset here so if next line panics/exits it is nil
			recording = recordExecution(rand.New(source), func(r *Runner) { checkable(G{r}, t) })
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
	var original []string
	for _, r := range executionRecord {
		original = append(original, fmt.Sprintf("%s: %s", r.Name(), i.config.Printer(r.Value())))
	}

	shrinks := 0
	for {
		shrank := false
		for ix := range executionRecord {
			didShrink := executionRecord[ix].AttemptShrink(func() bool {
				shrinks += 1
				return i.t.Run(fmt.Sprintf("%s (shrink attempt %v)", name, shrinks), func(t *testing.T) {
					replayExecution(executionRecord, func(r *Runner) { checkable(G{r}, t) })
				})
			})

			if didShrink {
				shrank = true
			}
		}

		if !shrank {
			break
		}
	}

	var report []string
	for _, r := range executionRecord {
		report = append(report, fmt.Sprintf("%s: %s", r.Name(), i.config.Printer(r.Value())))
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
