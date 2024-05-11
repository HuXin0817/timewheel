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
	BenchmarkTimeWheel
	BenchmarkTimeWheel-8   	       5	 200074033 ns/op
	BenchmarkStandard
	BenchmarkStandard-8    	       1	1000141833 ns/op
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
