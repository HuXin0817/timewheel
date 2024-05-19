package timewheel

import (
	"sync"
	"sync/atomic"
	"time"
)

const minInterval = 10 * time.Millisecond // Defines the minimum interval as 10 milliseconds

// timeslot struct represents a time slot, containing a callback function and a map for storing timers
type timeslot[T any] struct {
	do   func(elem *T)  // Callback function to be executed for each element
	slot map[int64][]*T // Map to store timers, keyed by their index
	mu   sync.Mutex     // Mutex for synchronizing access to the slot map
}

// Creates a new instance of timeslot
func newTimeslot[T any](do func(elem *T)) *timeslot[T] {
	return &timeslot[T]{
		do:   do,
		slot: make(map[int64][]*T),
	}
}

// Adds a timer to the timeslot at the specified index
func (ts *timeslot[T]) add(idx int64, t *T) *T {
	ts.mu.Lock()
	if s, ok := ts.slot[idx]; ok {
		ts.slot[idx] = append(s, t)
	} else {
		ts.slot[idx] = make([]*T, 1, 10) // Create a slice with initial capacity of 10
		ts.slot[idx][0] = t
	}
	ts.mu.Unlock()
	return t
}

// Executes the callback for all timers at the specified index and removes them from the slot
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

// Timer struct represents a single timer
type Timer struct {
	C      chan time.Time // Channel to signal when the timer fires
	belong *TimeWheel     // Reference to the TimeWheel to which this timer belongs
	stop   atomic.Bool    // Atomic boolean to indicate if the timer is stopped
}

// Stops the timer
func (t *Timer) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

// Ticker struct represents a repeating ticker
type Ticker struct {
	C         chan time.Time // Channel to signal when the ticker fires
	belong    *TimeWheel     // Reference to the TimeWheel to which this ticker belongs
	increment atomic.Int64   // Atomic integer to store the tick interval in terms of TimeWheel ticks
	stop      atomic.Bool    // Atomic boolean to indicate if the ticker is stopped
}

// Resets the ticker to fire at the specified duration
func (t *Ticker) Reset(d time.Duration) {
	t.increment.Store(int64(d / t.belong.interval))
}

// Stops the ticker
func (t *Ticker) Stop() {
	if t.stop.Load() {
		return
	}
	t.stop.Store(true)
	t.C <- t.belong.now
}

// TimeWheel struct represents the core of the timing wheel
type TimeWheel struct {
	now        time.Time         // Current time
	ticker     *time.Ticker      // Go Ticker to drive the TimeWheel
	current    int64             // Current tick index
	interval   time.Duration     // Duration of each tick
	timerSlot  *timeslot[Timer]  // Slot for managing timers
	tickerSlot *timeslot[Ticker] // Slot for managing tickers
	stop       atomic.Bool       // Atomic boolean to indicate if the TimeWheel is stopped
}

// Creates a new TimeWheel with the specified interval
func New(interval time.Duration) (tw *TimeWheel) {
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

// Returns the current time of the TimeWheel
func (tw *TimeWheel) Now() time.Time {
	return tw.now
}

// Returns the duration since the specified time
func (tw *TimeWheel) Since(t time.Time) time.Duration {
	return tw.now.Sub(t)
}

// Returns a channel that will receive the current time after the specified duration
func (tw *TimeWheel) After(d time.Duration) chan time.Time {
	return tw.NewTimer(d).C
}

// Stops the TimeWheel
func (tw *TimeWheel) Stop() {
	tw.stop.Store(true)
	tw.ticker.Stop()
}

// Resets the TimeWheel interval
func (tw *TimeWheel) Reset(d time.Duration) {
	tw.ticker.Reset(d)
	tw.interval = d
}

// Calculates the number of ticks for the specified duration
func (tw *TimeWheel) increment(d time.Duration) int64 {
	if d <= tw.interval {
		return 1
	}
	return int64(d / tw.interval)
}

// Creates a new timer that will fire after the specified duration
func (tw *TimeWheel) NewTimer(d time.Duration) *Timer {
	idx := tw.current + tw.increment(d)
	return tw.timerSlot.add(idx, &Timer{
		C:      make(chan time.Time, 1),
		belong: tw,
	})
}

// Creates a new ticker that will fire at the specified interval
func (tw *TimeWheel) NewTicker(d time.Duration) *Ticker {
	increment := tw.increment(d)
	ticker := &Ticker{
		C:      make(chan time.Time, 1),
		belong: tw,
	}
	ticker.increment.Store(increment)
	return tw.tickerSlot.add(tw.current+increment, ticker)
}
