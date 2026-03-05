package async

import (
	"testing"

	"go.k6.io/k6/js/modules"
)

func TestModuleInterface(t *testing.T) {
	var _ modules.Module = &RootModule{}
	var _ modules.Instance = &Async{}
}

func TestExports(t *testing.T) {
	a := &Async{}
	exports := a.Exports()
	if exports.Named == nil {
		t.Fatal("exports should have Named map")
	}
	if exports.Named["group"] == nil {
		t.Fatal("group should be exported")
	}
	if exports.Named["check"] == nil {
		t.Fatal("check should be exported")
	}
}
