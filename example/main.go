package main

import (
	"fmt"
	"github.com/HuXin0817/timewheel"
	"time"
)

func main() {
	tw := timewheel.New(500 * time.Millisecond)
	afterChan := tw.After(time.Second)
	timer := time.NewTimer(time.Second * 2)
	ticker := time.NewTicker(time.Second * 3)
	for range 5 {
		select {
		case <-afterChan:
			fmt.Print("afterChan out.")
		case <-ticker.C:
			fmt.Print("ticker tick.")
		case <-timer.C:
			fmt.Print("timer done.")
		}

		fmt.Println("at", tw.Now())
	}
}
