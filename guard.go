package guard

import "sync"

type locker struct {
	g *Guard
	c chan struct{}
}

func (l *locker) Lock() {
	if _, ok := <-l.c; !ok {
		panic("guard: duplicated locking")
	}
}

func (l *locker) Unlock() {
	l.g.unlock(l.c)
}

func (l *locker) Release() {
	go func() {
		l.Lock()
		l.Unlock()
	}()
}

// Guard grants exclusive access permission to clients.
type Guard struct {
	mu    sync.Mutex
	off   int
	locks []chan struct{}
}

func unblock(c chan struct{}) {
	c <- struct{}{}
	close(c)
}

func (g *Guard) unlock(c chan struct{}) {
	g.mu.Lock()
	l := len(g.locks)
	if g.off == l || g.locks[g.off] != c {
		panic("guard.Guard: unlocking unlocked")
	}
	g.off++
	switch {
	case g.off == l:
		g.off = 0
		g.locks = g.locks[:0]
	default:
		unblock(g.locks[g.off])
		if g.off >= l/2 || g.off >= 32 {
			n := l - g.off
			copy(g.locks, g.locks[g.off:l])
			g.off = 0
			g.locks = g.locks[:n]
		}
	}
	g.mu.Unlock()
}

// NewLocker creates a Locker for exclusive access permission acquisition.
// Lockers created after will not acquire the permission before this one got
// locked/unlocked or released.
func (g *Guard) NewLocker() Locker {
	c := make(chan struct{}, 1)
	g.mu.Lock()
	owned := g.off == len(g.locks)
	g.locks = append(g.locks, c)
	g.mu.Unlock()
	if owned {
		unblock(c)
	}
	return &locker{g: g, c: c}
}
