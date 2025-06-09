// Package runt provides a custom testing framework that mimics the standard
// testing package but allows for more flexible test execution outside of Go's
// built-in test runner. It supports context propagation, custom logging,
// and manual test orchestration.
package runt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// T represents a test instance that implements the testing.TB interface.
// It provides functionality similar to *testing.T but can be used outside
// of Go's standard test runner. T supports subtests, context propagation,
// logging, and proper error handling with panics.
type T struct {
	testing.TB
	name    string
	ctx     context.Context
	parent  *T
	logs    *strings.Builder
	failed  bool
	skipped bool
}

var _ testing.TB = (*T)(nil)

// New creates a new test instance with the given context and name.
// The context can be used to propagate cancellation and deadlines
// throughout the test execution.
func New(ctx context.Context, name string) *T {
	return &T{
		TB:   nil, // unused, has to be here because private()
		name: name,
		ctx:  ctx,
		logs: &strings.Builder{},
	}
}

// Name returns the name of the test.
func (e *T) Name() string {
	return e.name
}

// Run executes a subtest with the given name and callback function.
// It creates a new T instance for the subtest, handles panics gracefully,
// and returns true if the subtest passed (did not fail or skip).
// Any failure in the subtest will also mark the parent test as failed.
func (e *T) Run(name string, cb func(*T)) bool {
	sub := New(e.ctx, name)
	sub.parent = e
	// capture test panics, from assertions, skips, or otherwise
	defer func() {
		x := recover()
		switch x {
		case nil:
		case testSkipped{}, testFailed{}:
		default:
			sub.Errorf("PANIC: %v", x)
			sub.Error(debug.Stack())
		}
	}()
	cb(sub)
	return !sub.Failed()
}

// Helper marks the calling function as a test helper function.
// This implementation is a no-op for compatibility with testing.TB.
func (e *T) Helper() {}

// Logs returns all logged output from this test as a string.
// This includes output from Log, Logf, Error, Errorf, Fatal, and Fatalf calls.
func (e *T) Logs() string {
	return e.logs.String()
}

// Context returns the context associated with this test.
// The context can be used for cancellation and deadline propagation.
func (e *T) Context() context.Context {
	return e.ctx
}

// Error logs the arguments and marks the test as failed.
// The test will continue execution after calling Error.
func (e *T) Error(args ...any) {
	e.Log(args...)
	e.Fail()
}

// Errorf formats and logs the message and marks the test as failed.
// The test will continue execution after calling Errorf.
func (e *T) Errorf(format string, args ...any) {
	e.Logf(format, args...)
	e.Fail()
}

// Log logs the arguments to the test's log buffer.
// Arguments are handled similar to fmt.Println.
func (e *T) Log(args ...any) {
	fmt.Fprintln(e.logs, args...)
}

// Logf formats and logs the message to the test's log buffer.
// Format and arguments are handled similar to fmt.Printf.
func (e *T) Logf(format string, args ...any) {
	fmt.Fprintf(e.logs, format+"\n", args...)
}

// Fatal logs the arguments, marks the test as failed, and stops execution
// immediately by panicking. This will terminate the current test.
func (e *T) Fatal(args ...any) {
	e.Log(args...)
	e.FailNow()
}

// Fatalf formats and logs the message, marks the test as failed, and stops
// execution immediately by panicking. This will terminate the current test.
func (e *T) Fatalf(format string, args ...any) {
	e.Logf(format, args...)
	e.FailNow()
}

// Fail marks the test as failed but continues execution.
// If this test has a parent (i.e., it's a subtest), the parent is also marked as failed.
func (e *T) Fail() {
	if e.parent != nil {
		e.parent.Fail()
	}
	e.failed = true
}

type testFailed struct{}
type testSkipped struct{}

// FailNow marks the test as failed and stops execution immediately by panicking.
// This will terminate the current test but can be recovered by Run method.
func (e *T) FailNow() {
	e.failed = true
	panic(testFailed{})
}

// Failed returns true if the test has been marked as failed.
func (e *T) Failed() bool {
	return e.failed
}

// TempDir creates and returns a temporary directory for the test.
// The directory is created with a unique name based on the current timestamp.
// If directory creation fails, the test is terminated with Fatal.
func (e *T) TempDir() string {
	// Create temporary directory for test
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("evalT-%d", time.Now().UnixNano()))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		e.Fatal(err)
	}
	return dir
}

// Chdir changes the current working directory to the specified directory.
// If the directory change fails, the test is terminated with Fatal.
func (e *T) Chdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		e.Fatal(err)
	}
}

// Cleanup registers a cleanup function to be called when the test completes.
// This implementation is a no-op for compatibility with testing.TB.
func (e *T) Cleanup(func()) {}

// Setenv sets an environment variable for the duration of the test.
// If setting the environment variable fails, the test is terminated with Fatal.
func (e *T) Setenv(key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		e.Fatal(err)
	}
}

// Skip logs the arguments and marks the test as skipped.
// The test execution stops immediately.
func (e *T) Skip(args ...any) {
	e.Log(args...)
	e.SkipNow()
}

// Skipf formats and logs the message and marks the test as skipped.
// The test execution stops immediately.
func (e *T) Skipf(format string, args ...any) {
	e.Logf(format, args...)
	e.SkipNow()
}

// SkipNow marks the test as skipped and stops execution immediately by panicking.
// This will terminate the current test but can be recovered by Run method.
func (e *T) SkipNow() {
	e.skipped = true
	panic(testSkipped{})
}

// Skipped returns true if the test has been marked as skipped.
func (e *T) Skipped() bool {
	return e.skipped
}

// Deadline returns the deadline for the test execution, if any.
// It delegates to the underlying context's Deadline method.
func (e *T) Deadline() (time.Time, bool) {
	deadline, ok := e.ctx.Deadline()
	return deadline, ok
}
