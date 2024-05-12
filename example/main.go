package main

import (
	"fmt"
	"time"

	"github.com/HuXin0817/timewheel"
)

func main() {
	// create a new time wheel, interval is 500ms
	tw := timewheel.New(500 * time.Millisecond)
	defer tw.Stop()

	// create a after channel in 1s duration
	afterChan := tw.After(time.Second)

	// create a new timer in 2s duration
	timer := tw.NewTimer(time.Second * 2)

	// create a new ticker in 3s duration
	ticker := tw.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range 5 {
		select {
		case <-afterChan:
			fmt.Print("afterChan out.")
		case <-ticker.C:
			fmt.Print("ticker tick.")
		case <-timer.C:
			fmt.Print("timer done.")
			ticker.Reset(time.Second) // reset the interval of the ticker
		}

		fmt.Println("at", tw.Now())
	}
}
