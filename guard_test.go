package guard

import "testing"

func expectPanic(t *testing.T, err string) {
	r := recover()
	if r == nil {
		t.Fatalf("expect panic %q, got nothing", err)
	}
	switch got := r.(type) {
	case string:
		if got != err {
			t.Fatalf("expect panic %q, got %q", err, got)
		}
	default:
		t.Fatalf("expect panic %q, got %#v", err, got)
	}
}

func TestLockerLockLocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Lock()
	defer expectPanic(t, errLockerLocked)
	l.Lock()
}

func TestLockerLockUnlocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Lock()
	l.Unlock()
	defer expectPanic(t, errLockerUnlocked)
	l.Lock()
}

func TestLockerLockReleased(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Release()
	defer expectPanic(t, errLockerUnlocked)
	l.Lock()
}

func TestLockerUnlockNotLocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	defer expectPanic(t, errLockerNotLocked)
	l.Unlock()
}

func TestLockerUnlockUnlocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Lock()
	l.Unlock()
	defer expectPanic(t, errLockerUnlocked)
	l.Unlock()
}

func TestLockerUnlockReleased(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Release()
	defer expectPanic(t, errLockerUnlocked)
	l.Unlock()
}

func TestLockerReleaseLocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Lock()
	defer expectPanic(t, errLockerLocked)
	l.Release()
}

func TestLockerReleaseUnlocked(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Lock()
	l.Unlock()
	defer expectPanic(t, errLockerUnlocked)
	l.Release()
}

func TestLockerReleaseReleased(t *testing.T) {
	var g Guard
	l := g.NewLocker()
	l.Release()
	defer expectPanic(t, errLockerUnlocked)
	l.Release()
}
