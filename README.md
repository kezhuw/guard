# guard

[![GoDoc](https://godoc.org/github.com/kezhuw/guard?status.svg)](http://godoc.org/github.com/kezhuw/guard)
[![Build Status](https://travis-ci.org/kezhuw/guard.svg?branch=master)](https://travis-ci.org/kezhuw/guard)

package `guard` provides facilities to create lockers which can be locked in predetermined order.

## Usage
Guard and RWGuard can be used in one goroutine to create lockers, which can be
locked in other goroutines in creation order. This way the goroutine creating
lockers will not be blocked if permission is held by other goroutine, while
preserve acquisition orders of permission.

See [example_test.go](example_test.go) for example.

## License
The MIT License (MIT). See [LICENSE](LICENSE) for full license text.
