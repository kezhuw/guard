// Package guard provides facilities to create lockers
// which can be locked in predetermined order.
//
// Guard and RWGuard can be used in one goroutine to create lockers,
// which can be locked in other goroutines in creation order. This way
// the goroutine creating lockers will not be blocked if permission is
// held by other goroutine, while preserve acquisition orders of permission.
package guard
