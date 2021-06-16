# syncdo

## useage
```bash
package main

import (
	"fmt"
	"strconv"
	"time"
	"github.com/di3upham/syncdo"
)

func main() {
	for i := 0; i < 100; i++ {
		go func(index int) {
			lnum, k := new(int), strconv.Itoa(index%10) // do not reuse
			syncdo.Line(k, lnum, func() {
				time.Sleep(100 * time.Millisecond)
				fmt.Println(index)
				syncdo.Line(k, lnum, func() {}) // nested
			})
		}(i)
	}

	for i := 100; i < 200; i++ {
		go func(index int) {
			syncdo.Line(strconv.Itoa(index%5), new(int), func() {
				time.Sleep(100 * time.Millisecond)
				fmt.Println(index)
			})
		}(i)
	}

	for i := 200; i < 300; i++ {
		go func(index int) {
			lnum, k := new(int), strconv.Itoa(index%10) // do not reuse
			syncdo.KLock(k, lnum)
			time.Sleep(100 * time.Millisecond)
			fmt.Println(index)
			syncdo.KUnlock(k, lnum)
		}(i)
	}

	time.Sleep(1110 * time.Millisecond)
	fmt.Println(len(klm), len(kls), cap(kls))
}

```
