package guard

// Locker is a one-time sync.Locker, with additional Release method
// to abandon itself in initial state. It is designed for single
// goroutine usage.
type Locker interface {
	Lock()
	Unlock()

	// Release abandons locker, can only be called in initial state.
	// It starts a new goroutine to do Lock() and Unlock().
	Release()
}
