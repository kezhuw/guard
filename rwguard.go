package guard

import "sync"

const (
	errReaderLocked    = "guard.RWGuard: reader is already locked"
	errReaderUnlocked  = "guard.RWGuard: reader is already unlocked or released"
	errReaderNotLocked = "guard.RWGuard: reader is not locked"

	errWriterLocked    = "guard.RWGuard: writer is already locked"
	errWriterUnlocked  = "guard.RWGuard: writer is already unlocked or released"
	errWriterNotLocked = "guard.RWGuard: writer is not locked"
)

type reader struct {
	g *RWGuard
	c chan *RWGuard
}

func (r *reader) Lock() {
	c := r.c
	if c == nil {
		panic(errReaderUnlocked)
	}
	g, ok := <-c
	if !ok {
		panic(errReaderLocked)
	}
	r.g = g
}

func (r *reader) Unlock() {
	c := r.c
	switch {
	case c == nil:
		panic(errReaderUnlocked)
	case r.g == nil:
		panic(errReaderNotLocked)
	default:
		r.g.unlockRead(c)
		r.g = nil
		r.c = nil
	}
}

func (r *reader) Release() {
	c := r.c
	switch {
	case c == nil:
		panic(errReaderUnlocked)
	case r.g != nil:
		panic(errReaderLocked)
	default:
		go releaseRead(c)
		r.c = nil
	}
}

func releaseRead(c chan *RWGuard) {
	g, ok := <-c
	if !ok {
		panic("guard.RWGuard: reader is locked in releasing")
	}
	g.unlockRead(c)
}

type writer struct {
	g *RWGuard
	c chan *RWGuard
}

func (w *writer) Lock() {
	c := w.c
	if c == nil {
		panic(errWriterUnlocked)
	}
	g, ok := <-c
	if !ok {
		panic(errWriterLocked)
	}
	w.g = g
}

func (w *writer) Unlock() {
	c := w.c
	switch {
	case c == nil:
		panic(errWriterUnlocked)
	case w.g == nil:
		panic(errWriterNotLocked)
	default:
		w.g.unlockWrite(c)
		w.g = nil
		w.c = nil
	}
}

func (w *writer) Release() {
	c := w.c
	switch {
	case c == nil:
		panic(errWriterUnlocked)
	case w.g != nil:
		panic(errWriterLocked)
	default:
		go releaseWrite(c)
		w.c = nil
	}
}

func releaseWrite(c chan *RWGuard) {
	g, ok := <-c
	if !ok {
		panic("guard.RWGuard: writer is locked in releasing")
	}
	g.unlockWrite(c)
}

type waiter struct {
	c       chan *RWGuard
	writing bool
}

// RWGuard grants read and/or write permission to clients.
type RWGuard struct {
	mu      sync.Mutex
	off     int
	readers map[chan *RWGuard]struct{}
	waiters []waiter
}

func unblockRWGuardLocker(c chan *RWGuard, g *RWGuard) {
	c <- g
	close(c)
}

func (g *RWGuard) unlockRead(c chan *RWGuard) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.readers[c]; !ok {
		panic(errReaderNotLocked)
	}
	delete(g.readers, c)
	if len(g.readers) == 0 && g.off != len(g.waiters) {
		unblockRWGuardLocker(g.waiters[g.off].c, g)
	}
}

func (g *RWGuard) unlockWrite(c chan *RWGuard) {
	g.mu.Lock()
	defer g.mu.Unlock()
	i, l := g.off, len(g.waiters)
	if len(g.readers) != 0 || i == l || g.waiters[i].c != c {
		panic(errWriterNotLocked)
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
			unblockRWGuardLocker(c, g)
		default:
			unblockRWGuardLocker(c, g)
			g.waiters[i].c = nil
			g.readers[c] = struct{}{}
			i++
			for ; i < l; i++ {
				if g.waiters[i].writing {
					break
				}
				c := g.waiters[i].c
				unblockRWGuardLocker(c, g)
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
	c := make(chan *RWGuard, 1)
	g.mu.Lock()
	if g.readers == nil {
		g.readers = make(map[chan *RWGuard]struct{})
	}
	switch {
	case g.off == len(g.waiters):
		g.readers[c] = struct{}{}
		g.mu.Unlock()
		unblockRWGuardLocker(c, g)
	default:
		g.waiters = append(g.waiters, waiter{c: c})
		g.mu.Unlock()
	}
	return &reader{c: c}
}

// NewWriter creates a Locker for write permission acquisition.
// Lockers, including readers and writers, will not acquire their permission
// before this one got locked/unlocked or released.
func (g *RWGuard) NewWriter() Locker {
	c := make(chan *RWGuard, 1)
	g.mu.Lock()
	owned := len(g.readers) == 0 && len(g.waiters) == 0
	g.waiters = append(g.waiters, waiter{c: c, writing: true})
	g.mu.Unlock()
	if owned {
		unblockRWGuardLocker(c, g)
	}
	return &writer{c: c}
}
