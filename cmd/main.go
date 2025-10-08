package main

import (
	"fmt"
	"time"
)

func main() {
	loopNumber := 1
	for {
		fmt.Printf("dining-ordering, loop number v2: %d\n", loopNumber)
		loopNumber += 1
		time.Sleep(time.Second * 2)
	}
}
