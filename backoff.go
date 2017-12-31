package backoff

import (
	"math"
	"time"

	"sync"

	"github.com/go-playground/retry"
)

// RetryFn describes the retry function signature
type RetryFn = retry.Fn

// NotifyFn describes the notify function signature
type NotifyFn func(attempt uint16, waiting time.Duration, err error)

type averageWait struct {
	count       uint64
	accumulated float64
}

// Instance defines a backoff instance
type Instance struct {
	retries     uint16
	originalMin float64
	min         float64
	max         float64
	factor      float64
	m           sync.Mutex

	// auto adjust
	averageWait *averageWait
	wg          sync.WaitGroup
	done        chan struct{}
}

// New returns a new backoff instance for use with sane defaults
func New(retries uint16, min, max time.Duration) *Instance {
	i := &Instance{
		retries: retries,
		min:     float64(min),
		max:     float64(max),
	}
	i.originalMin = i.min
	i.calculateFactor()
	return i
}

// AutoTune automatically adjusts the minimum delay time based on past successes and failures to an acceptable rate.
// poll is the time interval after reset, or on initial setup, the value is calculated while reset is the time interval
// in which the values are set back to their original values and value will be recalculates after the poll duration.
func (i *Instance) AutoTune(poll, reset time.Duration) {
	i.averageWait = new(averageWait)
	i.done = make(chan struct{})
	i.wg.Add(1)
	go i.autoAdjust(poll, reset)
}

func (i *Instance) autoAdjust(poll, reset time.Duration) {
	t := time.NewTimer(poll)
	select {
	case <-i.done:
		if !t.Stop() {
			<-t.C
		}
		i.wg.Done()
		return
	case <-t.C:
	}
	i.m.Lock()
	if i.averageWait.count > 0 {
		i.min = i.averageWait.accumulated / float64(i.averageWait.count)
		i.calculateFactor()
	}
	i.m.Unlock()

	t.Reset(reset)
	select {
	case <-i.done:
		if !t.Stop() {
			<-t.C
		}
		i.wg.Done()
		return
	case <-t.C:
	}
	i.m.Lock()
	i.min = i.originalMin
	i.calculateFactor()
	i.m.Unlock()
	go i.autoAdjust(poll, reset)
}

// Run executes the provided function with exponential backoff upon failure
func (i *Instance) Run(fn RetryFn, notifyFn NotifyFn) error {
	notifyFunc := func(attempt uint16, err error) {
		i.m.Lock()
		f64 := i.min * math.Pow(i.factor, float64(attempt-1))
		if i.averageWait != nil {
			i.averageWait.count++
			i.averageWait.accumulated += f64
		}
		i.m.Unlock()
		wait := time.Duration(f64)
		if notifyFn != nil {
			notifyFn(attempt, wait, err)
		}
		time.Sleep(wait)
	}
	return retry.Run(i.retries, fn, notifyFunc)
}

func (i *Instance) calculateFactor() {
	i.factor = math.Pow(i.max/i.min, 1/float64(i.retries-1))
}

// Close cleans up any outstanding processes/goroutines
func (i *Instance) Close() {
	if i.done != nil {
		close(i.done)
	}
	i.wg.Wait()
}
