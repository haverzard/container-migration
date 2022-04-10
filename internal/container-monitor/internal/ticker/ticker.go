package ticker

import (
	"time"
)

const INTERVAL_PERIOD time.Duration = 24 * time.Hour

// Get the next trigger duration from current time
func getNextTickDuration(hours, minutes, seconds int) time.Duration {
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+hours, now.Minute()+minutes, now.Second()+seconds, 0, time.Local)
	if nextTick.Before(now) {
		nextTick = nextTick.Add(INTERVAL_PERIOD)
	}
	return nextTick.Sub(time.Now())
}

type JobTicker struct {
	T            *time.Timer
	HourToTick   int
	MinuteToTick int
	SecondToTick int
}

// Initialize job ticker
func NewJobTicker(hours, minutes, seconds int) *JobTicker {
	return &JobTicker{time.NewTimer(getNextTickDuration(hours, minutes, seconds)), hours, minutes, seconds}
}

// Set job ticker's next trigger
func (jt *JobTicker) UpdateJobTicker() {
	jt.T.Reset(getNextTickDuration(jt.HourToTick, jt.MinuteToTick, jt.SecondToTick))
}
