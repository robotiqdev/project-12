package health_test

import (
	"testing"
	"time"

	"github.com/robotiqdev/project-12/internal/health"
)

// TestNewTracker_StartTimeIsSetToNow verifies that after calling NewTracker,
// the tracker's StartTime equals the time returned by the injected now function.
func TestNewTracker_StartTimeIsSetToNow(t *testing.T) {
	fixed := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixed }

	tracker := health.NewTracker(nowFn)
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	got := tracker.StartTime()
	if !got.Equal(fixed) {
		t.Errorf("StartTime() = %v; want %v", got, fixed)
	}
}

// TestNewTracker_StartTimeIsApproximatelyNow verifies that when time.Now is
// passed as the now function, the recorded start time is within one second of
// the real current time (per the task spec: time.Since(tracker.StartTime()) < time.Second).
func TestNewTracker_StartTimeIsApproximatelyNow(t *testing.T) {
	before := time.Now()
	tracker := health.NewTracker(time.Now)
	after := time.Now()

	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	start := tracker.StartTime()

	if start.Before(before) {
		t.Errorf("StartTime() %v is before test start %v", start, before)
	}
	if start.After(after) {
		t.Errorf("StartTime() %v is after test end %v", start, after)
	}

	// Ensure it stays within the one-second window mandated by the task spec.
	if time.Since(start) >= time.Second {
		t.Errorf("time.Since(StartTime()) = %v; want < 1s", time.Since(start))
	}
}

// TestNewTracker_NowCalledExactlyOnce verifies that the now function is called
// exactly once during construction and not on subsequent StartTime() calls.
func TestNewTracker_NowCalledExactlyOnce(t *testing.T) {
	callCount := 0
	fixed := time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC)
	nowFn := func() time.Time {
		callCount++
		return fixed
	}

	tracker := health.NewTracker(nowFn)
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	if callCount != 1 {
		t.Errorf("now called %d times during NewTracker; want 1", callCount)
	}

	// Calling StartTime multiple times must not invoke now again.
	_ = tracker.StartTime()
	_ = tracker.StartTime()

	if callCount != 1 {
		t.Errorf("now called %d times total after StartTime() calls; want 1", callCount)
	}
}

// TestNewTracker_IndependentTrackers verifies that two trackers created with
// different now functions record independent start times.
func TestNewTracker_IndependentTrackers(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	tracker1 := health.NewTracker(func() time.Time { return t1 })
	tracker2 := health.NewTracker(func() time.Time { return t2 })

	if tracker1 == nil || tracker2 == nil {
		t.Fatal("NewTracker returned nil")
	}

	if !tracker1.StartTime().Equal(t1) {
		t.Errorf("tracker1.StartTime() = %v; want %v", tracker1.StartTime(), t1)
	}
	if !tracker2.StartTime().Equal(t2) {
		t.Errorf("tracker2.StartTime() = %v; want %v", tracker2.StartTime(), t2)
	}
}

// TestUptimeSeconds_ReturnsElapsedSeconds verifies that UptimeSeconds returns
// the number of seconds elapsed since the tracker was created, using a mutable
// now function that advances after construction.
func TestUptimeSeconds_ReturnsElapsedSeconds(t *testing.T) {
	fixedNow := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// nowTime starts at fixedNow; we advance it after the tracker is created.
	nowTime := fixedNow
	tracker := health.NewTracker(func() time.Time { return nowTime })

	// Advance by 5 seconds — subsequent calls to the injected now() return the new time.
	nowTime = fixedNow.Add(5 * time.Second)

	got := tracker.UptimeSeconds()
	if got != 5.0 {
		t.Errorf("UptimeSeconds() = %v; want 5.0", got)
	}
}

// TestUptimeSeconds_ZeroWhenNoTimeHasPassed verifies that UptimeSeconds returns
// 0.0 when the now function returns the same time as when the tracker was created.
func TestUptimeSeconds_ZeroWhenNoTimeHasPassed(t *testing.T) {
	fixedNow := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tracker := health.NewTracker(func() time.Time { return fixedNow })

	got := tracker.UptimeSeconds()
	if got != 0.0 {
		t.Errorf("UptimeSeconds() = %v; want 0.0", got)
	}
}

// TestNewTracker_StartTimeIsImmutable verifies that the start time does not
// change over the lifetime of the tracker (i.e., it is captured once and frozen).
func TestNewTracker_StartTimeIsImmutable(t *testing.T) {
	callCount := 0
	times := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	nowFn := func() time.Time {
		defer func() { callCount++ }()
		if callCount < len(times) {
			return times[callCount]
		}
		return times[len(times)-1]
	}

	tracker := health.NewTracker(nowFn)
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	first := tracker.StartTime()
	second := tracker.StartTime()

	if !first.Equal(second) {
		t.Errorf("StartTime() returned different values: %v vs %v", first, second)
	}
	if !first.Equal(times[0]) {
		t.Errorf("StartTime() = %v; want %v (first time returned by now)", first, times[0])
	}
}
