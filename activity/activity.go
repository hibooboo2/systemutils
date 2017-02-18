package activity

import (
	"errors"
	"time"
)

type ActivityMonitor interface {
	TimeSinceLastInput() (time.Duration, error)
	IsActive() bool
	IsInActive() bool
	UserActiveChan() chan struct{}
	UserInactiveChan() chan struct{}
}

type MonitorCreator func(activeThreshold, pollFrequency time.Duration) ActivityMonitor

var (
	NewActivityMonitor      MonitorCreator
	ErrFailedToGetLastInput = errors.New("Failed to get last user input")
)
