package runt

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRun_Success(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("This test should pass")
	})

	if err != nil {
		t.Errorf("Expected nil error for successful test, got: %v", err)
	}
}

func TestRun_Failure(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Error("This test should fail")
	})

	if err == nil {
		t.Fatal("Expected error for failed test, got nil")
	}

	if !strings.Contains(err.Error(), "test failed:") {
		t.Errorf("Expected error message to contain 'test failed:', got: %v", err)
	}

	if !strings.Contains(err.Error(), "This test should fail") {
		t.Errorf("Expected error message to contain test logs, got: %v", err)
	}
}

func TestRun_FailureWithFail(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("Some log message")
		tb.Fail()
	})

	if err == nil {
		t.Fatal("Expected error for failed test, got nil")
	}

	if !strings.Contains(err.Error(), "Some log message") {
		t.Errorf("Expected error message to contain test logs, got: %v", err)
	}
}

func TestRun_Skip(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Skip("This test should be skipped")
	})

	// Skipped tests should not return an error
	if err != nil {
		t.Errorf("Expected nil error for skipped test, got: %v", err)
	}
}

func TestRun_SkipWithSkipNow(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("Before skip")
		tb.SkipNow()
		tb.Log("This should not be logged")
	})

	// Skipped tests should not return an error
	if err != nil {
		t.Errorf("Expected nil error for skipped test, got: %v", err)
	}
}

func TestRun_Fatal(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Fatal("This should cause test failure")
	})

	if err == nil {
		t.Fatal("Expected error for fatal test, got nil")
	}

	if !strings.Contains(err.Error(), "This should cause test failure") {
		t.Errorf("Expected error message to contain fatal message, got: %v", err)
	}
}

func TestRun_Fatalf(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Fatalf("This should cause test failure: %s", "formatted")
	})

	if err == nil {
		t.Fatal("Expected error for fatal test, got nil")
	}

	if !strings.Contains(err.Error(), "This should cause test failure: formatted") {
		t.Errorf("Expected error message to contain formatted fatal message, got: %v", err)
	}
}

func TestRun_PanicWithTestFailed(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		panic(testFailed{})
	})

	// testFailed{} panic should be handled gracefully and not return an error
	// (unless the test was already marked as failed)
	if err != nil {
		t.Errorf("Expected nil error for test that panicked with testFailed{}, got: %v", err)
	}
}

func TestRun_PanicWithTestSkipped(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		panic(testSkipped{})
	})

	// testSkipped{} panic should be handled gracefully and not return an error
	if err != nil {
		t.Errorf("Expected nil error for test that panicked with testSkipped{}, got: %v", err)
	}
}

func TestRun_PanicWithOtherType(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		panic("unexpected panic")
	})

	if err == nil {
		t.Fatal("Expected error for test that panicked with unexpected type, got nil")
	}

	if !strings.Contains(err.Error(), "PANIC: unexpected panic") {
		t.Errorf("Expected error message to contain panic info, got: %v", err)
	}
}

func TestRun_PanicWithError(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")

	err := Run(ctx, func(tb testing.TB) {
		panic(testErr)
	})

	if err == nil {
		t.Fatal("Expected error for test that panicked with error, got nil")
	}

	if !strings.Contains(err.Error(), "PANIC: test error") {
		t.Errorf("Expected error message to contain panic info, got: %v", err)
	}
}

func TestRun_ContextPropagation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	err := Run(ctx, func(tb testing.TB) {
		runtT, ok := tb.(*T)
		if !ok {
			tb.Fatal("Expected *T type")
		}

		value := runtT.Context().Value("test-key")
		if value != "test-value" {
			tb.Errorf("Expected context value 'test-value', got: %v", value)
		}
	})

	if err != nil {
		t.Errorf("Expected nil error for context test, got: %v", err)
	}
}

func TestRun_MultipleOperations(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("First log")
		tb.Logf("Formatted log: %d", 42)
		tb.Error("First error")
		tb.Errorf("Formatted error: %s", "second")
	})

	if err == nil {
		t.Fatal("Expected error for failed test, got nil")
	}

	errorMsg := err.Error()
	expectedContent := []string{
		"First log",
		"Formatted log: 42",
		"First error",
		"Formatted error: second",
	}

	for _, content := range expectedContent {
		if !strings.Contains(errorMsg, content) {
			t.Errorf("Expected error message to contain '%s', got: %v", content, errorMsg)
		}
	}
}

func TestRun_TestName(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		if tb.Name() != "test" {
			tb.Errorf("Expected test name 'test', got: %s", tb.Name())
		}
	})

	if err != nil {
		t.Errorf("Expected nil error for name test, got: %v", err)
	}
}

func TestRun_LogsInErrorMessage(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("Debug info")
		tb.Logf("Value: %d", 123)
		tb.Error("Something went wrong")
	})

	if err == nil {
		t.Fatal("Expected error for failed test, got nil")
	}

	// Verify that all logs are included in the error message
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "Debug info") {
		t.Error("Expected error to contain 'Debug info'")
	}
	if !strings.Contains(errorMsg, "Value: 123") {
		t.Error("Expected error to contain 'Value: 123'")
	}
	if !strings.Contains(errorMsg, "Something went wrong") {
		t.Error("Expected error to contain 'Something went wrong'")
	}
}

func TestRun_SubtestExecution(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		runtT, ok := tb.(*T)
		if !ok {
			tb.Fatal("Expected *T type")
		}

		success := runtT.Run("subtest", func(sub *T) {
			sub.Log("Subtest log")
		})

		if !success {
			tb.Error("Expected subtest to succeed")
		}
	})

	if err != nil {
		t.Errorf("Expected nil error for subtest execution, got: %v", err)
	}
}

func TestRun_EmptyTest(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		// Empty test - should pass
	})

	if err != nil {
		t.Errorf("Expected nil error for empty test, got: %v", err)
	}
}

func TestRun_LogOnlyTest(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("Just logging, no failure")
		tb.Logf("Formatted log: %s", "test")
	})

	if err != nil {
		t.Errorf("Expected nil error for log-only test, got: %v", err)
	}
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := Run(ctx, func(tb testing.TB) {
		runtT, ok := tb.(*T)
		if !ok {
			tb.Fatal("Expected *T type")
		}

		// Verify context is cancelled
		select {
		case <-runtT.Context().Done():
			tb.Log("Context is cancelled as expected")
		default:
			tb.Error("Expected context to be cancelled")
		}
	})

	if err != nil {
		t.Errorf("Expected nil error for cancelled context test, got: %v", err)
	}
}

func TestRun_ErrorThenFatal(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Error("First error")
		tb.Fatal("Fatal error")
		tb.Log("This should not be reached")
	})

	if err == nil {
		t.Fatal("Expected error for test with Error then Fatal, got nil")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "First error") {
		t.Error("Expected error message to contain 'First error'")
	}
	if !strings.Contains(errorMsg, "Fatal error") {
		t.Error("Expected error message to contain 'Fatal error'")
	}
	if strings.Contains(errorMsg, "This should not be reached") {
		t.Error("Expected logs after Fatal to not be included")
	}
}

func TestRun_HelperCall(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Helper() // Should be a no-op but shouldn't cause issues
		tb.Log("Test with helper call")
	})

	if err != nil {
		t.Errorf("Expected nil error for test with Helper call, got: %v", err)
	}
}

func TestRun_MultipleSkips(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		tb.Log("Before skip")
		tb.Skip("First skip reason")
		tb.Skip("This should not be reached")
	})

	if err != nil {
		t.Errorf("Expected nil error for skipped test, got: %v", err)
	}
}

func TestRun_PanicThenRecover(t *testing.T) {
	ctx := context.Background()

	err := Run(ctx, func(tb testing.TB) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					tb.Logf("Recovered from panic: %v", r)
				}
			}()
			panic("test panic")
		}()
		tb.Log("After recovery")
	})

	if err != nil {
		t.Errorf("Expected nil error for test with internal panic recovery, got: %v", err)
	}
}
