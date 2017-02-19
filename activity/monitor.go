package activity

import (
	"errors"
	"time"
)

// Monitor interface to define the things used for user activity.
type Monitor interface {
	TimeSinceLastInput() (time.Duration, error)
	IsActive() bool
	IsInActive() bool
	UserActiveChan() chan struct{}
	UserInactiveChan() chan struct{}
}

//MonitorCreator a function that will creat an activity monitor.
type MonitorCreator func(activeThreshold, pollFrequency time.Duration) (Monitor, error)

var (
	//NewActivityMonitor function to get a new ActivityMonitor
	NewActivityMonitor      MonitorCreator
	ErrFailedToGetLastInput = errors.New("failed to get last user input")
)
