package tomb

import (
	"errors"
	"sync"
	"time"
)

var (
	WaitTimeoutError     = errors.New("timeout reached while waiting for Tomb to become dead")
	CannotStartDeadError = errors.New("cannot start goroutine in Tomb because it is dying or already dead")
)

// Tomb makes it easy to handle the lifecycle of goroutines.
// The Tomb has 3 states: Alive, Dying, Dead. It transitions in this order, and cannot go back to being Alive.
type Tomb struct {
	mut       sync.Mutex
	numAlive  int
	dyingChan chan int
	dying     bool
	deadChan  chan int
}

// Go starts the given function in a goroutine, unless the Tomb is Dying or Dead, in which case it will return an error.
// The function can read the Dying channel or check Alive to determine when to stop.
func (t *Tomb) Go(fn func()) error {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.dying {
		return CannotStartDeadError
	}

	t.ensureInitialized()

	t.numAlive++
	go func() {
		fn()
		t.mut.Lock()
		defer t.mut.Unlock()
		t.numAlive--
		t.tryCloseDead()
	}()

	return nil
}

// Kill the Tomb by setting it to the Dying state.
// Goroutines started with the Tomb receive the signal to stop.
// Kill does not wait for the goroutines to stop; for that the Wait or Dead functions should be used.
func (t *Tomb) Kill() {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.dying {
		return
	}

	t.ensureInitialized()
	t.dying = true
	close(t.dyingChan)

	t.tryCloseDead()
}

// Dying returns a channel that is closed when the Tomb reaches the Dying state.
// Goroutines started with Go can read from this channel to know when to stop, because the read unblocks
// when the channel is closed.
func (t *Tomb) Dying() chan int {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.ensureInitialized()
	return t.dyingChan
}

// Dead returns a channel that is closed when the Tomb reaches the Dead state.
// Users of the Tomb can read from this channel to know when all goroutines have stopped, because the read unblocks
// when the channel is closed.
func (t *Tomb) Dead() chan int {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.ensureInitialized()
	return t.deadChan
}

// Alive returns true when the Tomb is not Dying or Dead, and false otherwise.
//
// This function should only be used to check whether the Tomb is Dying or Dead,
// because otherwise there could be a situation where this function returns true
// and immediately after that the Tomb goes into Dying state.
func (t *Tomb) Alive() bool {
	t.mut.Lock()
	defer t.mut.Unlock()
	return !t.dying
}

// IsDead returns true when the Tomb has reached the Dead state.
func (t *Tomb) IsDead() bool {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.dying && t.numAlive == 0
}

// Wait blocks until the Tomb has reached the Dead state, or whenever the given timeout has been reached.
// When the timeout was reached, an error is returned.
func (t *Tomb) Wait(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		return WaitTimeoutError
	case <-t.Dead():
		// Shut down successfully.
		return nil
	}
}

func (t *Tomb) ensureInitialized() {
	if t.dyingChan == nil {
		t.dyingChan = make(chan int)
		t.deadChan = make(chan int)
	}
}

func (t *Tomb) tryCloseDead() {
	if t.dying && t.numAlive == 0 {
		close(t.deadChan)
	}
}
