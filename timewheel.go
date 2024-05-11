package timewheel

import (
	"sync"
	"sync/atomic"
	"time"
)

type timeslot[T any] struct {
	do   func(elem *T)
	slot map[int64][]*T
	mu   sync.Mutex
}

func newTimeslot[T any](do func(elem *T)) *timeslot[T] {
	return &timeslot[T]{
		do:   do,
		slot: make(map[int64][]*T),
	}
}

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

func (ts *timeslot[T]) done(idx int64) {
	if s, ok := ts.slot[idx]; ok {
		for _, t := range s {
			ts.do(t)
		}
		ts.mu.Lock()
		delete(ts.slot, idx)
		ts.mu.Unlock()
	}
}

type Timer struct {
	C      chan time.Time
	belong *TimeWheel
	stop   atomic.Bool
}

func (t *Timer) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

type Ticker struct {
	C      chan time.Time
	belong *TimeWheel
	addIdx int64
	stop   atomic.Bool
}

func (t *Ticker) Reset(d time.Duration) {
	t.addIdx = int64(d / t.belong.interval)
	if t.addIdx == 0 {
		t.addIdx = 1
	}
}

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

func (tw *TimeWheel) Now() time.Time { return tw.now }

func (tw *TimeWheel) Since(t time.Time) time.Duration { return tw.now.Sub(t) }

func (tw *TimeWheel) After(d time.Duration) chan time.Time { return tw.NewTimer(d).C }

func (tw *TimeWheel) Stop() {
	tw.stop.Store(true)
	tw.ticker.Stop()
}

func (tw *TimeWheel) Reset(d time.Duration) {
	tw.ticker.Reset(d)
	tw.interval = d
}

func (tw *TimeWheel) NewTimer(d time.Duration) (timer *Timer) {
	if d < tw.interval {
		d = tw.interval
	}
	return tw.timerSlot.add(tw.current+int64(d/tw.interval), &Timer{
		C:      make(chan time.Time, 1),
		belong: tw,
	})
}

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
