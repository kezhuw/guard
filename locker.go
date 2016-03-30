package guard

// Locker is a one-time sync.Locker, with additional Release method
// to abandon itself in initial state.
type Locker interface {
	Lock()
	Unlock()

	// Release abandons locker, can only be called in initial state.
	Release()
}
