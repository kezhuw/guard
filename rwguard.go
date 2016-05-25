package guard

import "sync"

type reader struct {
	g *RWGuard
	c chan struct{}
}

func (r *reader) Lock() {
	if _, ok := <-r.c; !ok {
		panic("guard.RWGuard: duplicated read locking")
	}
}

func (r *reader) Unlock() {
	r.g.unlockRead(r.c)
}

func (r *reader) Release() {
	go func() {
		r.Lock()
		r.Unlock()
	}()
}

type writer struct {
	g *RWGuard
	c chan struct{}
}

func (w *writer) Lock() {
	if _, ok := <-w.c; !ok {
		panic("guard.RWGuard: duplicated write locking")
	}
}

func (w *writer) Unlock() {
	w.g.unlockWrite(w.c)
}

func (w *writer) Release() {
	go func() {
		w.Lock()
		w.Unlock()
	}()
}

type waiter struct {
	c       chan struct{}
	writing bool
}

// RWGuard grants read and/or write permission to clients.
type RWGuard struct {
	mu      sync.Mutex
	off     int
	readers map[chan struct{}]struct{}
	waiters []waiter
}

func (g *RWGuard) unlockRead(c chan struct{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.readers[c]; !ok {
		panic("guard.RWGuard: unlocking unlocked reader")
	}
	delete(g.readers, c)
	if len(g.readers) == 0 && g.off != len(g.waiters) {
		unblock(g.waiters[g.off].c)
	}
}

func (g *RWGuard) unlockWrite(c chan struct{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	i, l := g.off, len(g.waiters)
	if len(g.readers) != 0 || i == l || g.waiters[i].c != c {
		panic("guard.RWGuard: unlocking unlocked writer")
	}
	g.waiters[i].c = nil
	i++
	switch {
	case i == l:
		g.off = 0
		g.waiters = g.waiters[:0]
	default:
		c := g.waiters[i].c
		switch g.waiters[i].writing {
		case true:
			unblock(c)
		default:
			unblock(c)
			g.waiters[i].c = nil
			g.readers[c] = struct{}{}
			i++
			for ; i < l; i++ {
				if g.waiters[i].writing {
					break
				}
				c := g.waiters[i].c
				unblock(c)
				g.waiters[i].c = nil
				g.readers[c] = struct{}{}
			}
			if i == l {
				g.off = 0
				g.waiters = g.waiters[:0]
				return
			}
		}
		switch {
		case i >= l/2 || i >= 32:
			copy(g.waiters, g.waiters[i:l])
			n := l - i
			collectWaiters(g.waiters[maxInt(i, n):l])
			g.off = 0
			g.waiters = g.waiters[:n]
		default:
			g.off = i
		}
	}
}

func collectWaiters(waiters []waiter) {
	for i, n := 0, len(waiters); i < n; i++ {
		waiters[i].c = nil
	}
}

// NewReader creates a Locker for read permission acquisition.
// Writers created after will not acquire their permission before this one got
// locked/unlocked or released.
func (g *RWGuard) NewReader() Locker {
	c := make(chan struct{}, 1)
	g.mu.Lock()
	if g.readers == nil {
		g.readers = make(map[chan struct{}]struct{})
	}
	switch {
	case g.off == len(g.waiters):
		g.readers[c] = struct{}{}
		g.mu.Unlock()
		unblock(c)
	default:
		g.waiters = append(g.waiters, waiter{c: c})
		g.mu.Unlock()
	}
	return &reader{g: g, c: c}
}

// NewWriter creates a Locker for write permission acquisition.
// Lockers, including readers and writers, will not acquire their permission
// before this one got locked/unlocked or released.
func (g *RWGuard) NewWriter() Locker {
	c := make(chan struct{}, 1)
	g.mu.Lock()
	owned := len(g.readers) == 0 && len(g.waiters) == 0
	g.waiters = append(g.waiters, waiter{c: c, writing: true})
	g.mu.Unlock()
	if owned {
		unblock(c)
	}
	return &writer{g: g, c: c}
}
