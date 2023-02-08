# A Safe Tomb Package
This package provides a safe way to do lifecycle management of goroutines.

In contrast to the more well-known [tomb package](https://github.com/go-tomb/tomb), this package has:

* no panics
* no [edge case causing Wait to block forever](https://github.com/go-tomb/tomb/issues/21)
* simpler API
* does not kill itself after the last goroutine has finished, unless the Tomb has been killed

## Example

```go
package example

import (
	"fmt"
	"github.com/milandamen/tomb"
	"time"
)

func DoWork() error {
	var t tomb.Tomb
	err := t.Go(func() {
		timer := time.NewTimer(time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("hello world 1")
			case <-t.Dying():
				return
			}
		}
	})
	if err != nil {
		return err
	}

	err = t.Go(func() {
		timer := time.NewTimer(time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("hello world 2")
			case <-t.Dying():
				return
			}
		}
	})
	if err != nil {
		return err
	}

	// Do other work
	time.Sleep(5 * time.Second)

	// Stop the goroutines using the tomb.
	t.Kill()
	err = t.Wait(2 * time.Second)
	if err != nil {
		return err
	}

	return nil
}

```

## License
This project is licensed under the terms of the [MIT license](LICENSE).

