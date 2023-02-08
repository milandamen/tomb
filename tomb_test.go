package tomb

import (
	"testing"
	"time"
)

func TestTomb_HappyFlow(t1 *testing.T) {
	var t Tomb
	var b bool
	var c bool
	err := t.Go(func() {
		b = true
		<-t.Dying()
	})
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Go(func() {
		c = true
		<-t.Dying()
	})
	if err != nil {
		t1.Fatal(err)
	}

	t.Kill()
	err = t.Wait(time.Second)
	if err != nil {
		t1.Fatal(err)
	}

	if !b {
		t1.Fatalf("goroutine b was not run")
	}
	if !c {
		t1.Fatalf("goroutine c was not run")
	}
}

func TestTomb_NoGoroutineStartedBeforeKill(t1 *testing.T) {
	var t Tomb
	t.Kill()
	err := t.Wait(time.Second)
	if err != nil {
		t1.Fatal(err)
	}
}

func TestTomb_StartGoroutineWhileDead(t1 *testing.T) {
	var t Tomb
	t.Kill()
	err := t.Wait(time.Second)
	if err != nil {
		t1.Fatal(err)
	}

	err = t.Go(func() {})
	if err == nil {
		t1.Fatal("expected cannot start goroutine error but got no error")
	}
	if err != CannotStartDeadError {
		t1.Fatal("expected cannot start goroutine error but got: " + err.Error())
	}
}

func TestTomb_StartGoroutineWhileDying(t1 *testing.T) {
	var t Tomb
	waitChan := make(chan int)
	err := t.Go(func() {
		<-waitChan
	})
	if err != nil {
		t1.Fatal(err)
	}

	t.Kill()

	err = t.Go(func() {})
	if err == nil {
		t1.Fatal("expected cannot start goroutine error but got no error")
	}
	if err != CannotStartDeadError {
		t1.Fatal("expected cannot start goroutine error but got: " + err.Error())
	}

	close(waitChan)

	err = t.Wait(time.Second)
	if err != nil {
		t1.Fatal(err)
	}
}

func TestTomb_StuckGoroutine(t1 *testing.T) {
	var t Tomb
	waitChan := make(chan int)
	defer close(waitChan)

	err := t.Go(func() {
		<-waitChan
	})
	if err != nil {
		t1.Fatal(err)
	}

	t.Kill()
	err = t.Wait(time.Second)
	if err == nil {
		t1.Fatal("expected timeout but got no error")
	}
	if err != WaitTimeoutError {
		t1.Fatal("expected timeout error but got: " + err.Error())
	}
}
