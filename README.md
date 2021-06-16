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
	unlock := syncdo.KLock("f1")
	fmt.Println("hello world")
	unlock()
}

func fx() {
	defer syncdo.KLock("f1")()
	// function body
}

```
