package activity

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32           = syscall.MustLoadDLL("user32.dll")
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	getLastInputInfo = user32.MustFindProc("GetLastInputInfo")
	getTickCount     = kernel32.MustFindProc("GetTickCount")
)

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

// Watcher used to encapsulate the activity functionality so it can be configured.
type Watcher struct {
	//ActiveThreshold period of time before a user is considered inactive.
	activeThreshold time.Duration
	//PollFrequency how often channels are reported user activity.
	pollFrequency    time.Duration
	active, inactive []chan struct{}
	li               lastInputInfo
	sync.Mutex
}

// NewWatcher returns a watcher that can be used to get user activity.
func newWindowsWatcher(activeThreshold, pollFrequency time.Duration) ActivityMonitor {
	w := Watcher{}
	w.li.cbSize = uint32(unsafe.Sizeof(w.li))
	w.activeThreshold = activeThreshold
	w.pollFrequency = pollFrequency
	go w.timeUpdater()
	go w.activityPoller()
	return &w
}

func (w *Watcher) TimeSinceLastInput() (time.Duration, error) {
	tick, _, _ := getTickCount.Call()
	return time.Duration(uint32(tick)-w.li.dwTime) * time.Millisecond, nil
}

func (w *Watcher) timeUpdater() {
	for {
		time.Sleep(w.pollFrequency)
		w.Lock()
		r1, _, _ := getLastInputInfo.Call(uintptr(unsafe.Pointer(&w.li)))
		w.Unlock()
		if r1 == 0 {
			continue
		}
	}
}

func (w *Watcher) IsActive() bool {
	t, _ := w.TimeSinceLastInput()
	return t < w.activeThreshold
}

func (w *Watcher) IsInActive() bool {
	t, _ := w.TimeSinceLastInput()
	return t > w.activeThreshold
}

// UserActiveChan returns a chan that when an empty struct is received
// means that a user is active.
func (w *Watcher) UserActiveChan() chan struct{} {
	w.Lock()
	a := make(chan struct{})
	w.active = append(w.active, a)
	w.Unlock()
	return a
}

// activityPoller checks to see if the user is inactive or active if it is
func (w *Watcher) activityPoller() {
	var active bool
	var previousActive bool
	for {
		time.Sleep(w.pollFrequency)
		active = w.IsActive()
		if active != previousActive {
			switch {
			case active:
				for _, a := range w.active {
					a <- struct{}{}
				}
			case !active:
				for _, a := range w.inactive {
					a <- struct{}{}
				}
			}
		}
		previousActive = active
	}
}

// UserInActiveChan returns a chan that when an empty struct is received
// means that a user is inactive.
func (w *Watcher) UserInactiveChan() chan struct{} {
	a := make(chan struct{})
	w.Lock()
	w.inactive = append(w.inactive, a)
	w.Unlock()
	return a
}

func init() {
	NewActivityMonitor = newWindowsWatcher
}
