package async

import (
	"testing"
)

func TestCheckSync(t *testing.T) {
	rt, _ := setupTest(t)

	val, err := rt.RunString(`
		check(null, {
			"test1": true,
			"test2": function() { return true; }
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	if !val.ToBoolean() {
		t.Fatal("expected check to pass")
	}
}

func TestCheckSyncFail(t *testing.T) {
	rt, _ := setupTest(t)

	val, err := rt.RunString(`
		check(null, {
			"test1": true,
			"test2": false
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	if val.ToBoolean() {
		t.Fatal("expected check to fail")
	}
}

func TestCheckAsync(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"test1": async function() { return true; },
			"test2": async function() { return true; }
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckMixed(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"sync_test": true,
			"async_test": async function() { return true; },
			"func_test": function() { return true; }
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckWithArg(t *testing.T) {
	rt, _ := setupTest(t)

	val, err := rt.RunString(`
		var obj = { status: 200 };
		check(obj, {
			"status is 200": function(r) { return r.status === 200; }
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	if !val.ToBoolean() {
		t.Fatal("expected check to pass")
	}
}

func TestCheckNoChecks(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, null);
	`)
	if err == nil {
		t.Fatal("expected error for null checks")
	}
}

func TestCheckInitContext(t *testing.T) {
	rt, vu := setupTest(t)
	vu.StateField = nil

	_, err := rt.RunString(`
		check(null, {"test": true});
	`)
	if err == nil {
		t.Fatal("expected error in init context")
	}
}

func TestCheckMetrics(t *testing.T) {
	rt, vu := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"metric_test1": true,
			"metric_test2": false
		});
	`)
	if err != nil {
		t.Fatal(err)
	}

	if len(vu.StateField.Samples) == 0 {
		t.Fatal("expected check metrics")
	}
}

func TestCheckWithTags(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"test": true
		}, {
			"custom_tag": "value"
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckGroupSeparator(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"test::invalid": true
		});
	`)
	if err == nil {
		t.Fatal("expected error for group separator in check name")
	}
}

func TestCheckFunctionError(t *testing.T) {
	rt, _ := setupTest(t)

	_, err := rt.RunString(`
		check(null, {
			"error_test": function() { throw new Error("test error"); }
		});
	`)
	if err == nil {
		t.Fatal("expected error from check function")
	}
}
