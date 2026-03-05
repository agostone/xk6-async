package async

import (
	"errors"
	"strings"
	"time"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
)

func (a *Async) Check(arg0, checks sobek.Value, extras ...sobek.Value) (sobek.Value, error) {
	state := a.vu.State()
	if state == nil {
		return nil, errors.New("using check() in the init context is not supported")
	}
	if checks == nil {
		return nil, errors.New("no checks provided to `check`")
	}

	runtime := a.vu.Runtime()
	t := time.Now()

	commonTagsAndMeta := state.Tags.GetCurrentValues()
	if len(extras) > 0 {
		if err := common.ApplyCustomUserTags(runtime, &commonTagsAndMeta, extras[0]); err != nil {
			return nil, err
		}
	}

	obj := checks.ToObject(runtime)
	keys := obj.Keys()
	results := make([]checkResult, len(keys))
	hasAsync := false

	for i, name := range keys {
		if strings.Contains(name, lib.GroupSeparator) {
			return nil, lib.ErrNameContainsGroupSeparator
		}

		val := obj.Get(name)
		results[i] = checkResult{name: name, tags: commonTagsAndMeta}

		fn, ok := sobek.AssertFunction(val)
		if ok {
			tmpVal, err := fn(sobek.Undefined(), arg0)
			if err != nil {
				results[i].value = runtime.ToValue(false)
				results[i].err = err
			} else {
				results[i].value = tmpVal
				if tmpVal != nil && tmpVal.ExportType().String() == "*sobek.Promise" {
					hasAsync = true
				}
			}
		} else {
			results[i].value = val
		}
	}

	if hasAsync {
		return a.handleAsyncChecks(results, state, t)
	}

	return a.emitCheckResults(results, state, t)
}

type checkResult struct {
	name  string
	value sobek.Value
	tags  metrics.TagsAndMeta
	err   error
}

func (a *Async) handleAsyncChecks(results []checkResult, state *lib.State, t time.Time) (sobek.Value, error) {
	runtime := a.vu.Runtime()
	promises := make([]interface{}, 0)

	for _, result := range results {
		if result.value != nil && result.value.ExportType().String() == "*sobek.Promise" {
			promises = append(promises, result.value)
		}
	}

	if len(promises) == 0 {
		return a.emitCheckResults(results, state, t)
	}

	promiseAll := runtime.GlobalObject().Get("Promise").ToObject(runtime).Get("all")
	allFn, ok := sobek.AssertFunction(promiseAll)
	if !ok {
		return a.emitCheckResults(results, state, t)
	}

	promiseArray := runtime.NewArray(promises...)
	combinedPromise, err := allFn(sobek.Undefined(), promiseArray)
	if err != nil || combinedPromise == nil {
		return a.emitCheckResults(results, state, t)
	}

	thenFn := runtime.ToValue(func(sobek.FunctionCall) sobek.Value {
		if _, err := a.emitCheckResults(results, state, t); err != nil {
			a.vu.State().Logger.WithError(err).Error("failed to emit check results")
		}
		return runtime.ToValue(true)
	})

	catchFn := runtime.ToValue(func(sobek.FunctionCall) sobek.Value {
		if _, err := a.emitCheckResults(results, state, t); err != nil {
			a.vu.State().Logger.WithError(err).Error("failed to emit check results")
		}
		return runtime.ToValue(false)
	})

	promiseObj := combinedPromise.ToObject(runtime)
	thenMethod := promiseObj.Get("then")
	if thenMethod == nil {
		return a.emitCheckResults(results, state, t)
	}

	thenFunc, ok := sobek.AssertFunction(thenMethod)
	if !ok {
		return a.emitCheckResults(results, state, t)
	}

	resultPromise, err := thenFunc(combinedPromise, thenFn)
	if err != nil || resultPromise == nil {
		return a.emitCheckResults(results, state, t)
	}

	catchMethod := resultPromise.ToObject(runtime).Get("catch")
	if catchMethod != nil {
		if catchFunc, ok := sobek.AssertFunction(catchMethod); ok {
			return catchFunc(resultPromise, catchFn)
		}
	}

	return resultPromise, nil
}

func (a *Async) emitCheckResults(results []checkResult, state *lib.State, t time.Time) (sobek.Value, error) {
	ctx := a.vu.Context()
	success := true

	for _, result := range results {
		if result.err != nil {
			return a.vu.Runtime().ToValue(false), result.err
		}

		booleanVal := result.value.ToBoolean()
		if !booleanVal {
			success = false
		}

		tags := result.tags.Tags
		if state.Options.SystemTags.Has(metrics.TagCheck) {
			tags = tags.With("check", result.name)
		}

		sample := metrics.Sample{
			TimeSeries: metrics.TimeSeries{
				Metric: state.BuiltinMetrics.Checks,
				Tags:   tags,
			},
			Time:     t,
			Metadata: result.tags.Metadata,
		}
		if booleanVal {
			sample.Value = 1
		}

		metrics.PushIfNotDone(ctx, state.Samples, sample)
	}

	return a.vu.Runtime().ToValue(success), nil
}
