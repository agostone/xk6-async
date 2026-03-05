package async

import (
	"errors"
	"time"

	"github.com/grafana/sobek"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
)

func (a *Async) Group(name string, val sobek.Value) (sobek.Value, error) {
	state := a.vu.State()
	if state == nil {
		return nil, errors.New("using group() in the init context is not supported")
	}

	fn, ok := sobek.AssertFunction(val)
	if !ok {
		return nil, errors.New("group() requires a callback as a second argument")
	}

	oldGroupName, _, shouldUpdateTag, err := a.setupGroup(state, name)
	if err != nil {
		return sobek.Undefined(), err
	}

	startTime := time.Now()
	ret, err := fn(sobek.Undefined())

	if ret != nil && ret.ExportType().String() == "*sobek.Promise" {
		return a.handleAsync(ret, state, startTime, oldGroupName, shouldUpdateTag)
	}

	return a.handleSync(ret, err, state, startTime, oldGroupName, shouldUpdateTag)
}

func (a *Async) setupGroup(state *lib.State, name string) (string, string, bool, error) {
	oldGroupName, _ := state.Tags.GetCurrentValues().Tags.Get(metrics.TagGroup.String())
	newGroupName, err := lib.NewGroupPath(oldGroupName, name)
	if err != nil {
		return "", "", false, err
	}

	shouldUpdateTag := state.Options.SystemTags.Has(metrics.TagGroup)
	if shouldUpdateTag {
		state.Tags.Modify(func(tagsAndMeta *metrics.TagsAndMeta) {
			tagsAndMeta.SetSystemTagOrMeta(metrics.TagGroup, newGroupName)
		})
	}

	return oldGroupName, newGroupName, shouldUpdateTag, nil
}

func (a *Async) handleAsync(ret sobek.Value, state *lib.State, startTime time.Time, oldGroupName string, shouldUpdateTag bool) (sobek.Value, error) {
	runtime := a.vu.Runtime()

	finallyFn := runtime.ToValue(func(sobek.FunctionCall) sobek.Value {
		a.emitMetrics(state, startTime)
		if shouldUpdateTag {
			state.Tags.Modify(func(tagsAndMeta *metrics.TagsAndMeta) {
				tagsAndMeta.SetSystemTagOrMeta(metrics.TagGroup, oldGroupName)
			})
		}
		return sobek.Undefined()
	})

	if finallyMethod := ret.ToObject(runtime).Get("finally"); finallyMethod != nil {
		if fn, ok := sobek.AssertFunction(finallyMethod); ok {
			fn(ret, finallyFn)
		}
	}
	return ret, nil
}

func (a *Async) handleSync(ret sobek.Value, err error, state *lib.State, startTime time.Time, oldGroupName string, shouldUpdateTag bool) (sobek.Value, error) {
	a.emitMetrics(state, startTime)

	if shouldUpdateTag {
		state.Tags.Modify(func(tagsAndMeta *metrics.TagsAndMeta) {
			tagsAndMeta.SetSystemTagOrMeta(metrics.TagGroup, oldGroupName)
		})
	}

	return ret, err
}

func (a *Async) emitMetrics(state *lib.State, startTime time.Time) {
	t := time.Now()
	ctx := a.vu.Context()
	ctm := state.Tags.GetCurrentValues()
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: state.BuiltinMetrics.GroupDuration,
			Tags:   ctm.Tags,
		},
		Time:     t,
		Value:    metrics.D(t.Sub(startTime)),
		Metadata: ctm.Metadata,
	})
}
