# TimeWheel

A lightweight time wheel implemented in golang.

## Install

To install timewheel, input command:

```sh
$ go get -u github.com/HuXin0817/timewheel
```

## Usage

You can use timewheel like the golang time package, it contains the methods like `Now`, `Since`, `After`, etc.

```go
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
```

## Benchmark

follow data is timewheel compare with standard time package:

    goos: darwin
    goarch: arm64
    pkg: github.com/HuXin0817/timewheel
    BenchmarkTimeWheel
    BenchmarkTimeWheel-8   	       5	 200074033 ns/op
    BenchmarkStandard
    BenchmarkStandard-8    	       1	1000141833 ns/op
    PASS

## Contributions

Contributions are welcome! If you find any issues or have suggestions for improvements, feel free to open an issue or create a pull request on the GitHub repository.