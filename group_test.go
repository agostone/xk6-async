package async

import (
	"testing"
)

func TestGroupSync(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		var called = false;
		group("test", function() {
			called = true;
			return 42;
		});
		if (!called) throw new Error("callback not called");
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupAsync(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		var called = false;
		group("test", async function() {
			called = true;
			return 42;
		});
		if (!called) throw new Error("callback not called");
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupInvalidCallback(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		group("test", "not a function");
	`)
	if err == nil {
		t.Fatal("expected error for non-function callback")
	}
}

func TestGroupInitContext(t *testing.T) {
	rt, vu := setupTest(t)
	vu.StateField = nil

	_, err := rt.RunString(`
		group("test", function() {});
	`)
	if err == nil {
		t.Fatal("expected error in init context")
	}
}

func TestGroupMetrics(t *testing.T) {
	rt, vu := setupTest(t)

	_, err := rt.RunString(`
		group("metrics_test", function() {
			return "done";
		});
	`)
	if err != nil {
		t.Fatal(err)
	}

	if len(vu.StateField.Samples) == 0 {
		t.Fatal("expected group_duration metric")
	}
}
