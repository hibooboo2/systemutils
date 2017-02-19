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
func newWindowsWatcher(activeThreshold, pollFrequency time.Duration) (Monitor, error) {
	w := Watcher{}
	w.li.cbSize = uint32(unsafe.Sizeof(w.li))
	w.activeThreshold = activeThreshold
	w.pollFrequency = pollFrequency
	_, err := w.TimeSinceLastInput()
	if err != nil {
		return nil, err
	}
	go w.timeUpdater()
	go w.activityPoller()
	return &w, nil
}

func (w *Watcher) TimeSinceLastInput() (time.Duration, error) {
	w.Lock()
	r1, _, _ := getLastInputInfo.Call(uintptr(unsafe.Pointer(&w.li)))
	w.Unlock()
	if r1 == 0 {
		return 0, ErrFailedToGetLastInput
	}
	tick, _, _ := getTickCount.Call()
	return time.Duration(uint32(tick)-w.li.dwTime) * time.Millisecond, nil
}

func init() {
	NewActivityMonitor = newWindowsWatcher
}
