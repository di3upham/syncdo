package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	for i := 0; i < 100; i++ {
		go func(index int) {
			lnum := new(int) // do not reuse
			k := strconv.Itoa(index % 10)
			Line(k, lnum, func() {
				time.Sleep(100 * time.Millisecond)
				Line(k, lnum, func() {}) // nested
			})
		}(i)
	}

	for i := 0; i < 100; i++ {
		go func(index int) { Line(strconv.Itoa(index%5), new(int), func() { time.Sleep(100 * time.Millisecond) }) }(i)
	}

	time.Sleep(1110 * time.Millisecond)
	fmt.Println(len(klm), len(kls), cap(kls))
}
