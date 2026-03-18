package health

import "time"

// Tracker holds health-tracking state including when tracking began.
type Tracker struct {
	startTime time.Time
	now       func() time.Time
}

// NewTracker creates a new Tracker using the provided time function.
// The now function is called once at construction to record the start time.
// Pass time.Now for production use; pass a fixed function for testing.
func NewTracker(now func() time.Time) *Tracker {
	return nil // stub — implementation not yet provided
}

// StartTime returns the time at which the Tracker was created.
func (t *Tracker) StartTime() time.Time {
	return time.Time{}
}
