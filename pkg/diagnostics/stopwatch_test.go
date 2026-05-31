package diagnostics

import (
	"testing"
	"time"
)

func TestStopwatchReportsElapsedTime(t *testing.T) {
	var stopwatch Stopwatch
	stopwatch.Reset()

	time.Sleep(time.Millisecond)

	if stopwatch.Microseconds() <= 0 {
		t.Fatal("expected positive microseconds")
	}
	if stopwatch.Milliseconds() < 0 {
		t.Fatal("milliseconds should not be negative")
	}
	if stopwatch.Seconds() <= 0 {
		t.Fatal("expected positive seconds")
	}
}
