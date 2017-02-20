package activity

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Watcher used to encapsulate the activity functionality so it can be configured.
type Watcher struct {
	//ActiveThreshold period of time before a user is considered inactive.
	activeThreshold time.Duration
	//PollFrequency how often channels are reported user activity.
	pollFrequency    time.Duration
	active, inactive []chan struct{}
	sync.Mutex
}

// NewWatcher returns a watcher that can be used to get user activity.
func newDarwinWatcher(activeThreshold, pollFrequency time.Duration) (Monitor, error) {
	w := Watcher{}
	w.activeThreshold = activeThreshold
	w.pollFrequency = pollFrequency
	_, err := w.TimeSinceLastInput()
	if err != nil {
		return nil, err
	}
	go w.activityPoller()
	return &w, nil
}

func (w *Watcher) TimeSinceLastInput() (time.Duration, error) {
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	o, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	// o := run("cat", "ioreg")
	const idleTime = `"HIDIdleTime" = `
	const idleN = len(idleTime)
	n := strings.Index(string(o), idleTime)
	cutOff := n + 50
	if cutOff > len(o) {
		cutOff = len(o) - 1
	}
	s := strings.Index(string(o[n:cutOff]), "\n")
	o = o[n+idleN : n+s]
	i, err := strconv.ParseInt(string(o), 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(i/1000000) * time.Millisecond, nil
}

func init() {
	NewActivityMonitor = newDarwinWatcher
}
