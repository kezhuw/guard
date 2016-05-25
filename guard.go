package guard

import "sync"

const (
	errLockerLocked    = "guard.Guard: locker is already locked"
	errLockerUnlocked  = "guard.Guard: locker is already unlocked or released"
	errLockerNotLocked = "guard.Guard: locker is not locked"
)

type locker struct {
	g *Guard
	c chan *Guard
}

func (l *locker) Lock() {
	c := l.c
	if c == nil {
		panic(errLockerUnlocked)
	}
	g, ok := <-c
	if !ok {
		panic(errLockerLocked)
	}
	l.g = g
}

func (l *locker) Unlock() {
	c := l.c
	switch {
	case c == nil:
		panic(errLockerUnlocked)
	case l.g == nil:
		panic(errLockerNotLocked)
	default:
		l.g.unlock(c)
		l.g = nil
		l.c = nil
	}
}

func (l *locker) Release() {
	c := l.c
	switch {
	case c == nil:
		panic(errLockerUnlocked)
	case l.g != nil:
		panic(errLockerLocked)
	default:
		go releaseLocker(c)
		l.c = nil
	}
}

func releaseLocker(c chan *Guard) {
	g, ok := <-c
	if !ok {
		panic("guard.Guard: locker is locked in releasing")
	}
	g.unlock(c)
}

// Guard grants exclusive access permission to clients.
type Guard struct {
	mu    sync.Mutex
	off   int
	locks []chan *Guard
}

func unblockGuardLocker(c chan *Guard, g *Guard) {
	c <- g
	close(c)
}

func (g *Guard) unlock(c chan *Guard) {
	g.mu.Lock()
	l := len(g.locks)
	if g.off == l || g.locks[g.off] != c {
		panic(errLockerNotLocked)
	}
	g.locks[g.off] = nil
	g.off++
	switch {
	case g.off == l:
		g.off = 0
		g.locks = g.locks[:0]
	default:
		unblockGuardLocker(g.locks[g.off], g)
		if g.off > maxHoleOffset {
			copy(g.locks, g.locks[g.off:l])
			n := l - g.off
			collectLockers(g.locks[maxInt(g.off, n):l])
			g.off = 0
			g.locks = g.locks[:n]
		}
	}
	g.mu.Unlock()
}

func collectLockers(lockers []chan *Guard) {
	for i, n := 0, len(lockers); i < n; i++ {
		lockers[i] = nil
	}
}

// NewLocker creates a Locker for exclusive access permission acquisition.
// Lockers created after will not acquire the permission before this one got
// locked/unlocked or released.
func (g *Guard) NewLocker() Locker {
	c := make(chan *Guard, 1)
	g.mu.Lock()
	owned := g.off == len(g.locks)
	g.locks = append(g.locks, c)
	g.mu.Unlock()
	if owned {
		unblockGuardLocker(c, g)
	}
	return &locker{c: c}
}
