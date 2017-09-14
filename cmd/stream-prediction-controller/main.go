package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("stream-prediction-controller")
	timerEvents := time.Tick(time.Second)
	for t := range timerEvents {
		fmt.Printf("time: %v\n", t)
	}
}
