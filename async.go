package async

import (
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/async", &RootModule{})
}

type RootModule struct{}

type Async struct {
	vu modules.VU
}

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Async{}
)

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Async{vu: vu}
}

func (a *Async) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"group": a.Group,
			"check": a.Check,
		},
	}
}
