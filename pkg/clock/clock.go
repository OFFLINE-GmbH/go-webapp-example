package clock

import "time"

// Clock is a wrapper around time functions to make
// time dependent code easily testable.
type Clock struct {
	now time.Time
}

// FromTime returns a new clock with a set time.
func FromTime(t time.Time) *Clock {
	return &Clock{now: t}
}

// Return the current time.
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

// Set the current time.
func (c *Clock) SetNow(t time.Time) {
	c.now = t
}
