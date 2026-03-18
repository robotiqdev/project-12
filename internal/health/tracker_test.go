package health

import (
	"testing"
	"time"
)

// TestNewTracker_StartTimeIsSetCorrectly verifies that NewTracker records the
// time returned by the injected now function as startTime.
func TestNewTracker_StartTimeIsSetCorrectly(t *testing.T) {
	fixed := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixed }

	tracker := NewTracker(nowFn)
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	if got, want := tracker.startTime, fixed; !got.Equal(want) {
		t.Errorf("startTime = %v; want %v", got, want)
	}
}

// TestUptimeSeconds_ZeroElapsed verifies that UptimeSeconds returns 0.0 when
// the now function returns the same time as at construction.
func TestUptimeSeconds_ZeroElapsed(t *testing.T) {
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := NewTracker(func() time.Time { return fixed })

	if got, want := tracker.UptimeSeconds(), 0.0; got != want {
		t.Errorf("UptimeSeconds() = %v; want %v", got, want)
	}
}

// TestUptimeSeconds_FiveSecondsElapsed verifies that UptimeSeconds returns 5.0
// when the now function returns T0+5s after construction at T0.
func TestUptimeSeconds_FiveSecondsElapsed(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	nowTime := t0
	tracker := NewTracker(func() time.Time { return nowTime })

	nowTime = t0.Add(5 * time.Second)

	if got, want := tracker.UptimeSeconds(), 5.0; got != want {
		t.Errorf("UptimeSeconds() = %v; want %v", got, want)
	}
}

// TestUptimeSeconds_FractionalSeconds verifies that UptimeSeconds correctly
// returns fractional values (e.g., 1.5 seconds).
func TestUptimeSeconds_FractionalSeconds(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	nowTime := t0
	tracker := NewTracker(func() time.Time { return nowTime })

	nowTime = t0.Add(1500 * time.Millisecond)

	if got, want := tracker.UptimeSeconds(), 1.5; got != want {
		t.Errorf("UptimeSeconds() = %v; want %v", got, want)
	}
}
