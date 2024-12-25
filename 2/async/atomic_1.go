package main

import (
	"fmt"
	"time"
)

var totalOperations_ int32 = 0

func inc_() {
	totalOperations_++
}

func main() {
	for i := 0; i < 1000; i++ {
		go inc_()
	}
	time.Sleep(20 * time.Millisecond)
	// ождается 1000, но по факту будет меньше
	fmt.Println("total operation = ", totalOperations_)
}
