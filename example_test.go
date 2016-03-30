package guard_test

import (
	"fmt"
	"time"

	"github.com/kezhuw/guard"
)

func ExampleGuard() {
	var guard guard.Guard
	l0 := guard.NewLocker()
	l1 := guard.NewLocker()
	go func() {
		time.Sleep(time.Millisecond * 500)
		l0.Lock()
		fmt.Println("l0 locked.")
		l0.Unlock()
	}()
	l1.Lock()
	fmt.Println("l1 locked.")
	l1.Unlock()
	// Output:
	// l0 locked.
	// l1 locked.
}

func ExampleRWGuard() {
	var guard guard.RWGuard
	r0 := guard.NewReader()
	r1 := guard.NewReader()
	w0 := guard.NewWriter()
	r0.Lock()
	done := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 500)
		r1.Lock()
		fmt.Println("r1 locked.")
		r1.Unlock()
	}()
	go func() {
		w0.Lock()
		fmt.Println("w0 locked.")
		w0.Unlock()
		close(done)
	}()
	fmt.Println("r0 locked.")
	r0.Unlock()
	<-done
	// Output:
	// r0 locked.
	// r1 locked.
	// w0 locked.
}
