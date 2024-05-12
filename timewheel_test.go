package timewheel_test

import (
	"testing"
	"time"

	"github.com/HuXin0817/timewheel"
)

/*
	goos: darwin
	goarch: arm64
	pkg: github.com/HuXin0817/timewheel
	BenchmarkTimeWheel-8   	       8	 125043838 ns/op	     727 B/op	       7 allocs/op
	BenchmarkStandard-8    	       1	1000140750 ns/op	    4552 B/op	      31 allocs/op
	PASS
*/

const interval = time.Second

var tw = timewheel.New(500 * time.Millisecond)

func BenchmarkTimeWheel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			<-tw.After(interval)
		}
	})
}

func BenchmarkStandard(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			<-time.After(interval)
		}
	})
}
