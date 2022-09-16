# 𓁦 maat 𓆄 

maat is an experiment

----

Write a test that fails:

```go
package test

import (
	"testing"

	"github.com/Porges/maat/go"
	"github.com/Porges/maat/go/gen"
	"github.com/kr/pretty"
)

func TestWithMaat(t *testing.T) {
	opts := maat.PrettyPrinter(func(value any) string { return pretty.Sprint(value) })
	m := maat.New(t, opts)

	m.Boolean("slice elements are equal",
		func(g maat.G) bool {
			slice := maat.SliceOf[int](g, "values", gen.Int, 2, 100)
			return slice[0] == slice[1]
		})
}
```

Run it:

```
/workspaces/maat/test/maat_impl.go:23:
    [Maat] Shrunk failure:
    values: []int{0, 1}

    [Maat] Original failure:
    values: []int{8153555194712150324, 5924735709383197940, 8849436958677232277, 7617028578081934146, 6611937046398741256, 8918980849700278997, 543753657066021060, 5381539986688807447, 6849128387853601179, 8879276983217379127, 4688226476298360828, 83262479907313568, 6646628255543796755, 8872856915208710320, 2191487501422702678, 7623574017229396991, 9120882265996367109, 3551821174991089734, 6612791590976887188, 8842529859930961400, 3324417932911324715, 6969073557920120438, 6892087633253323955, 2363317189431784511, 8314820358479786150, 225031222956847629, 2915981562036269880, 8933856656932910348}
```

Thanks `maat`.
