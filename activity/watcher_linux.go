package activity

import (
	"sync"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"
)

// Watcher used to encapsulate the activity functionality so it can be configured.
type Watcher struct {
	//ActiveThreshold period of time before a user is considered inactive.
	activeThreshold time.Duration
	//PollFrequency how often channels are reported user activity.
	pollFrequency    time.Duration
	active, inactive []chan struct{}
	root             xproto.Window
	xConn            *xgb.Conn
	sync.Mutex
}

// NewWatcher returns a watcher that can be used to get user activity.
//This will start one go routine.
func newLinuxWatcher(activeThreshold, pollFrequency time.Duration) (Monitor, error) {
	w := Watcher{}
	w.activeThreshold = activeThreshold
	w.pollFrequency = pollFrequency

	X, err := xgb.NewConn()
	w.xConn = X
	if err != nil {
		return nil, err
	}
	if err = screensaver.Init(X); err != nil {
		return nil, err
	}
	w.root = xproto.Setup(X).DefaultScreen(X).Root
	_, err = w.TimeSinceLastInput()
	if err != nil {
		return nil, err
	}
	go w.activityPoller()
	return &w, nil
}

func (w *Watcher) TimeSinceLastInput() (time.Duration, error) {
	info, err := screensaver.QueryInfo(w.xConn, xproto.Drawable(w.root)).Reply()
	if err != nil {
		return 0, ErrFailedToGetLastInput
	}
	return time.Duration(info.MsSinceUserInput) * time.Millisecond, nil
}

func init() {
	NewActivityMonitor = newLinuxWatcher
}
