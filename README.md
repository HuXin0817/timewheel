# TimeWheel

A lightweight time wheel implemented in Go.

## Install

To install TimeWheel, use the following command:

```sh
$ go get -u github.com/HuXin0817/timewheel
```

## Usage

TimeWheel can be used similarly to the standard Go `time` package. It includes methods such as `Now`, `Since`, `After`, and more.

### Example

```go
package main

import (
	"fmt"
	"time"
	"github.com/HuXin0817/timewheel"
)

func main() {
	// Create a new TimeWheel with a 500-millisecond interval
	tw := timewheel.New(500 * time.Millisecond)
	defer tw.Stop() // Ensure the TimeWheel stops when the main function exits

	// Create a channel that will receive a time event after 1 second
	afterChan := tw.After(time.Second)

	// Create a timer that will fire after 2 seconds
	timer := tw.NewTimer(time.Second * 2)

	// Create a ticker that will tick every 3 seconds
	ticker := tw.NewTicker(time.Second * 3)
	defer ticker.Stop() // Ensure the ticker stops when the main function exits

	// Run the loop 5 times
	for i := 0; i < 5; i++ {
		select {
		case <-afterChan:
			fmt.Print("afterChan out.") // Print message when afterChan fires
		case <-ticker.C:
			fmt.Print("ticker tick.") // Print message when the ticker ticks
		case <-timer.C:
			fmt.Print("timer done.") // Print message when the timer is done
			ticker.Reset(time.Second) // Reset the ticker to tick every 1 second
		}

		// Print the current time from the TimeWheel
		fmt.Println("at", tw.Now())
	}
}
```

## Benchmark

The following data compares TimeWheel with the standard time package:

```
goos: darwin
goarch: arm64
pkg: github.com/HuXin0817/timewheel
BenchmarkTimeWheel
BenchmarkTimeWheel-8           5   200074033 ns/op
BenchmarkStandard
BenchmarkStandard-8            1  1000141833 ns/op
PASS
```

## Contributions

Contributions are welcome! If you find any issues or have suggestions for improvements, feel free to open an issue or create a pull request on the [GitHub repository](https://github.com/HuXin0817/timewheel).
