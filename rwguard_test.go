package guard

import "testing"

func TestReaderLockLocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Lock()
	defer expectPanic(t, errReaderLocked)
	r.Lock()
}

func TestReaderLockUnlocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Lock()
	r.Unlock()
	defer expectPanic(t, errReaderUnlocked)
	r.Lock()
}

func TestReaderLockReleased(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Release()
	defer expectPanic(t, errReaderUnlocked)
	r.Lock()
}

func TestReaderUnlockNotLocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	defer expectPanic(t, errReaderNotLocked)
	r.Unlock()
}

func TestReaderUnlockUnlocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Lock()
	r.Unlock()
	defer expectPanic(t, errReaderUnlocked)
	r.Unlock()
}

func TestReaderUnlockReleased(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Release()
	defer expectPanic(t, errReaderUnlocked)
	r.Unlock()
}

func TestReaderReleaseLocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Lock()
	defer expectPanic(t, errReaderLocked)
	r.Release()
}

func TestReaderReleaseUnlocked(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Lock()
	r.Unlock()
	defer expectPanic(t, errReaderUnlocked)
	r.Release()
}

func TestReaderReleaseReleased(t *testing.T) {
	var g RWGuard
	r := g.NewReader()
	r.Release()
	defer expectPanic(t, errReaderUnlocked)
	r.Release()
}

func TestWriterLockLocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Lock()
	defer expectPanic(t, errWriterLocked)
	w.Lock()
}

func TestWriterLockUnlocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Lock()
	w.Unlock()
	defer expectPanic(t, errWriterUnlocked)
	w.Lock()
}

func TestWriterLockReleased(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Release()
	defer expectPanic(t, errWriterUnlocked)
	w.Lock()
}

func TestWriterUnlockNotLocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	defer expectPanic(t, errWriterNotLocked)
	w.Unlock()
}

func TestWriterUnlockUnlocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Lock()
	w.Unlock()
	defer expectPanic(t, errWriterUnlocked)
	w.Unlock()
}

func TestWriterUnlockReleased(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Release()
	defer expectPanic(t, errWriterUnlocked)
	w.Unlock()
}

func TestWriterReleaseLocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Lock()
	defer expectPanic(t, errWriterLocked)
	w.Release()
}

func TestWriterReleaseUnlocked(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Lock()
	w.Unlock()
	defer expectPanic(t, errWriterUnlocked)
	w.Release()
}

func TestWriterReleaseReleased(t *testing.T) {
	var g RWGuard
	w := g.NewWriter()
	w.Release()
	defer expectPanic(t, errWriterUnlocked)
	w.Release()
}
