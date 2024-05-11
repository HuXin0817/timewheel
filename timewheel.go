package timewheel

import (
	"sync"
	"sync/atomic"
	"time"
)

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
		ts.slot[idx] = []*T{t}
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
	C      chan time.Time
	belong *TimeWheel
	addIdx int64
	stop   atomic.Bool
}

// Reset reset the interval of the ticker
func (t *Ticker) Reset(d time.Duration) { t.addIdx = int64(d / t.belong.interval) }

// Stop stop the ticker and block the channel
func (t *Ticker) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

type TimeWheel struct {
	now        time.Time
	ticker     *time.Ticker
	current    int64
	interval   time.Duration
	timerSlot  *timeslot[Timer]
	tickerSlot *timeslot[Ticker]
	stop       atomic.Bool
}

// New Create a time wheel for a specified interval
func New(interval time.Duration) (tw *TimeWheel) {
	if interval < time.Millisecond {
		interval = time.Millisecond
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
			tw.tickerSlot.add(tw.current+t.addIdx, t)
		}),
	}

	go func() {
		for tw.now = range tw.ticker.C {
			if tw.stop.Load() {
				return
			}
			tw.current++
			idx := tw.current
			tw.timerSlot.done(idx)
			tw.tickerSlot.done(idx)
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

// Reset the interval of timewheel
func (tw *TimeWheel) Reset(d time.Duration) {
	tw.ticker.Reset(d)
	tw.interval = d
}

// NewTimer Create a new timer with timeout
func (tw *TimeWheel) NewTimer(d time.Duration) (timer *Timer) {
	if d < tw.interval {
		d = tw.interval
	}
	return tw.timerSlot.add(tw.current+int64(d/tw.interval), &Timer{
		C:      make(chan time.Time, 1),
		belong: tw,
	})
}

// NewTicker Create a new ticker with interval
func (tw *TimeWheel) NewTicker(d time.Duration) (ticker *Ticker) {
	if d < tw.interval {
		d = tw.interval
	}
	return tw.tickerSlot.add(tw.current+int64(d/tw.interval), &Ticker{
		C:      make(chan time.Time, 1),
		belong: tw,
		addIdx: int64(d / tw.interval),
	})
}
