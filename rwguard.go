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

func (rw *RWGuard) unlockRead(c chan struct{}) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if _, ok := rw.readers[c]; !ok {
		panic("guard.RWGuard: unlocking unlocked reader")
	}
	delete(rw.readers, c)
	if len(rw.readers) == 0 && rw.off != len(rw.waiters) {
		unblock(rw.waiters[rw.off].c)
	}
}

func (rw *RWGuard) unlockWrite(c chan struct{}) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	i, l := rw.off, len(rw.waiters)
	if len(rw.readers) != 0 || i == l || rw.waiters[i].c != c {
		panic("guard.RWGuard: unlocking unlocked writer")
	}
	rw.waiters[i].c = nil
	i++
	switch {
	case i == l:
		rw.off = 0
		rw.waiters = rw.waiters[:0]
	default:
		c := rw.waiters[i].c
		switch rw.waiters[i].writing {
		case true:
			unblock(c)
		default:
			unblock(c)
			rw.waiters[i].c = nil
			rw.readers[c] = struct{}{}
			i++
			for ; i < l; i++ {
				if rw.waiters[i].writing {
					break
				}
				c := rw.waiters[i].c
				unblock(c)
				rw.waiters[i].c = nil
				rw.readers[c] = struct{}{}
			}
			if i == l {
				rw.off = 0
				rw.waiters = rw.waiters[:0]
				return
			}
		}
		switch {
		case i >= l/2 || i >= 32:
			copy(rw.waiters, rw.waiters[i:l])
			n := l - i
			if n > i {
				i = n
			}
			for ; i < l; i++ {
				rw.waiters[i].c = nil
			}
			rw.off = 0
			rw.waiters = rw.waiters[:n]
		default:
			rw.off = i
		}
	}
}

// NewReader creates a Locker for read permission acquisition.
// Writers created after will not acquire their permission before this one got
// locked/unlocked or released.
func (rw *RWGuard) NewReader() Locker {
	c := make(chan struct{}, 1)
	rw.mu.Lock()
	if rw.readers == nil {
		rw.readers = make(map[chan struct{}]struct{})
	}
	switch {
	case rw.off == len(rw.waiters):
		rw.readers[c] = struct{}{}
		rw.mu.Unlock()
		unblock(c)
	default:
		rw.waiters = append(rw.waiters, waiter{c: c})
		rw.mu.Unlock()
	}
	return &reader{g: rw, c: c}
}

// NewWriter creates a Locker for write permission acquisition.
// Lockers, including readers and writers, will not acquire their permission
// before this one got locked/unlocked or released.
func (rw *RWGuard) NewWriter() Locker {
	c := make(chan struct{}, 1)
	rw.mu.Lock()
	owned := len(rw.readers) == 0 && len(rw.waiters) == 0
	rw.waiters = append(rw.waiters, waiter{c: c, writing: true})
	rw.mu.Unlock()
	if owned {
		unblock(c)
	}
	return &writer{g: rw, c: c}
}
