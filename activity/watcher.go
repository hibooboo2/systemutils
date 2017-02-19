package activity

import "time"

// IsActive returns true if user is active.
func (w *Watcher) IsActive() bool {
	t, err := w.TimeSinceLastInput()
	if err != nil {
		return false
	}
	return t < w.activeThreshold
}

// IsInActive returns true if user is inactive.
func (w *Watcher) IsInActive() bool {
	t, err := w.TimeSinceLastInput()
	if err != nil {
		return true
	}
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

// UserInactiveChan returns a chan that when an empty struct is received
//means that a user is inactive.
func (w *Watcher) UserInactiveChan() chan struct{} {
	a := make(chan struct{})
	w.Lock()
	w.inactive = append(w.inactive, a)
	w.Unlock()
	return a
}
