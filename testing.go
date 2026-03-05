package async

import (
	"context"
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
	"testing"
)

func setupTest(t *testing.T) (*sobek.Runtime, *modulestest.VU) {
	runtime := sobek.New()
	runtime.SetFieldNameMapper(common.FieldNameMapper{})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	root := &RootModule{}
	registry := metrics.NewRegistry()

	samples := make(chan metrics.SampleContainer, 1000)
	state := &lib.State{
		Options:        lib.Options{},
		BufferPool:     lib.NewBufferPool(),
		Samples:        samples,
		Tags:           lib.NewVUStateTags(registry.RootTagSet()),
		BuiltinMetrics: metrics.RegisterBuiltinMetrics(registry),
	}

	mockVU := &modulestest.VU{
		RuntimeField: runtime,
		StateField:   state,
		CtxField:     ctx,
	}

	moduleInstance := root.NewModuleInstance(mockVU)
	exports := moduleInstance.Exports().Named
	if err := runtime.Set("group", exports["group"]); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Set("check", exports["check"]); err != nil {
		t.Fatal(err)
	}

	return runtime, mockVU
}
