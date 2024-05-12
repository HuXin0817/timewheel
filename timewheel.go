package timewheel

import (
	"sync"
	"sync/atomic"
	"time"
)

const minInterval = 10 * time.Millisecond // the min interval, not surport smaller than it

// timeslot is a struct storage and deal the elements in the slot
type timeslot[T any] struct {
	do   func(elem *T)
	slot map[int64][]*T
	mu   sync.Mutex
}

// newTimeslot Create a new timeslot by function that deal one element
func newTimeslot[T any](do func(elem *T)) *timeslot[T] {
	return &timeslot[T]{
		do:   do,
		slot: make(map[int64][]*T),
	}
}

// add Append the element into the corresponding slot
func (ts *timeslot[T]) add(idx int64, t *T) *T {
	ts.mu.Lock()
	if s, ok := ts.slot[idx]; ok {
		ts.slot[idx] = append(s, t)
	} else {
		// make slice with cap 10
		ts.slot[idx] = make([]*T, 1, 10)
		ts.slot[idx][0] = t
	}
	ts.mu.Unlock()
	return t
}

// done Deal and clear all the element in corresponding slot
func (ts *timeslot[T]) done(idx int64) {
	if s, ok := ts.slot[idx]; ok {
		for _, t := range s {
			ts.do(t)
		}
		ts.mu.Lock()
		delete(ts.slot, idx) // clear the slot after done
		ts.mu.Unlock()
	}
}

// Timer a suitable timer struct in timeWheel
type Timer struct {
	C      chan time.Time
	belong *TimeWheel
	stop   atomic.Bool
}

// Stop stop the timer and block the channel
func (t *Timer) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

// Ticker A suitable ticker struct in timeWheel
type Ticker struct {
	C         chan time.Time
	belong    *TimeWheel
	increment atomic.Int64
	stop      atomic.Bool
}

// Reset reset the interval of the ticker
func (t *Ticker) Reset(d time.Duration) { t.increment.Store(int64(d / t.belong.interval)) }

// Stop stop the ticker and block the channel
func (t *Ticker) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

type TimeWheel struct {
	now        time.Time         // the time wheel now time
	ticker     *time.Ticker      // interval ticker
	current    int64             // the current duration
	interval   time.Duration     // the timewheel interval
	timerSlot  *timeslot[Timer]  // timer slot
	tickerSlot *timeslot[Ticker] // ticker slot
	stop       atomic.Bool       // stop signal
}

// New Create a time wheel for a specified interval
func New(interval time.Duration) (tw *TimeWheel) {
	// it the interval given smaller than the min interval, fix it to the mininterval
	if interval < minInterval {
		interval = minInterval
	}

	tw = &TimeWheel{
		now:       time.Now(),
		interval:  interval,
		ticker:    time.NewTicker(interval),
		timerSlot: newTimeslot(func(t *Timer) { t.Stop() }),
		tickerSlot: newTimeslot(func(t *Ticker) {
			if t.stop.Load() {
				return
			}
			t.C <- tw.now
			tw.tickerSlot.add(tw.current+t.increment.Load(), t)
		}),
	}

	go func() {
		for tw.now = range tw.ticker.C {
			if tw.stop.Load() {
				return
			}
			tw.current++
			tw.timerSlot.done(tw.current)
			tw.tickerSlot.done(tw.current)
		}
	}()

	return
}

// Now return the NowTime in the time wheel, accuracy is within a time interval
func (tw *TimeWheel) Now() time.Time { return tw.now }

// Now return the NowTime in the time wheel, accuracy is within a time interval
func (tw *TimeWheel) Since(t time.Time) time.Duration { return tw.now.Sub(t) }

// After Create a channel, and auto fill it in duration
func (tw *TimeWheel) After(d time.Duration) chan time.Time { return tw.NewTimer(d).C }

// Stop the time wheel
func (tw *TimeWheel) Stop() {
	tw.stop.Store(true)
	tw.ticker.Stop()
}

// Reset the time interval of timewheel
func (tw *TimeWheel) Reset(d time.Duration) {
	tw.ticker.Reset(d)
	tw.interval = d
}

// increment transform the duration to the index increment
func (tw *TimeWheel) increment(d time.Duration) int64 {
	if d <= tw.interval {
		return 1
	}
	return int64(d / tw.interval)
}

// NewTimer Create a new timer with timeout
func (tw *TimeWheel) NewTimer(d time.Duration) (timer *Timer) {
	idx := tw.current + tw.increment(d)
	return tw.timerSlot.add(idx, &Timer{
		C:      make(chan time.Time, 1),
		belong: tw,
	})
}

// NewTicker Create a new ticker with interval
func (tw *TimeWheel) NewTicker(d time.Duration) (ticker *Ticker) {
	increment := tw.increment(d)
	ticker = &Ticker{
		C:      make(chan time.Time, 1),
		belong: tw,
	}
	ticker.increment.Store(increment)
	return tw.tickerSlot.add(tw.current+increment, ticker)
}
