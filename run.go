package runt

import (
	"context"
	"fmt"
	"testing"
)

// Run executes a test function with a new test instance.
// It creates a new T instance with the provided context and executes the
// provided callback function. The function handles panics gracefully and
// returns an error if the test fails or is skipped, including the test logs.
// Returns nil if the test passes successfully.
func Run(ctx context.Context, cb func(testing.TB)) error {
	t := New(ctx, "test")

	(func() {
		// capture test panics, from assertions, skips, or otherwise
		defer func() {
			x := recover()
			switch x {
			case nil:
			case testSkipped{}, testFailed{}:
			default:
				t.Errorf("PANIC: %v", x)
			}
		}()
		cb(t)
	})()

	if t.Failed() {
		return fmt.Errorf("test failed:\n%s", t.Logs())
	}

	return nil
}
