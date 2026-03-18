package health

import "time"

// Tracker tracks server health metrics such as uptime.
type Tracker struct {
	startTime time.Time
}

// NewTracker creates a new Tracker with the current time as start.
func NewTracker() *Tracker {
	return &Tracker{startTime: time.Now()}
}

// UptimeSeconds returns the number of seconds since the Tracker was created.
func (t *Tracker) UptimeSeconds() float64 {
	return time.Since(t.startTime).Seconds()
}
