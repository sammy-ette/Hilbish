package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sammy-ette/hilbish/moonlight"
)

type timerType int64

const (
	timerInterval timerType = iota
	timerTimeout
)

// @type
// @interface timers
// @since 2.0.0
// @property type What kind of timer it is: interval (repeating) or timeout (one-shot).
// @property running Whether the timer is currently running.
// @property duration The duration in milliseconds after which the callback fires.
// @property id The ID of the timer.
// The Timer type represents a Hilbish timer created with hilbish.timers.create.
type timer struct {
	mu      sync.Mutex
	id      int
	typ     timerType
	running bool
	dur     time.Duration
	fun     *moonlight.Closure
	th      *timersModule
	ticker  *time.Ticker
	ud      *moonlight.UserData
	channel chan struct{}
}

func (t *timer) start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return errors.New("timer is already running")
	}

	if t.dur <= 0 {
		return errors.New("timer duration must be positive")
	}

	t.running = true
	t.th.mu.Lock()
	t.th.running++
	t.th.mu.Unlock()
	t.th.wg.Add(1)
	t.ticker = time.NewTicker(t.dur)

	go func() {
		for {
			select {
			case <-t.ticker.C:
				_, err := l.Call1(moonlight.FunctionValue(t.fun))
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error in function:\n", err)
					t.stop()
				}
				// only run one for timeout
				if t.typ == timerTimeout {
					t.stop()
				}
			case <-t.channel:
				t.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (t *timer) stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return errors.New("timer not running")
	}

	t.channel <- struct{}{}
	t.running = false
	t.th.mu.Lock()
	t.th.running--
	t.th.mu.Unlock()
	t.th.wg.Done()

	return nil
}

// @interface timers
// @member
// start()
// Starts a timer.
// @since 2.0.0
func timerStart(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	t, err := timerArg(mlr, 0)
	if err != nil {
		return err
	}

	err = t.start()
	if err != nil {
		return err
	}

	return nil
}

// @interface timers
// @member
// stop()
// Stops a timer.
// @since 2.0.0
func timerStop(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	t, err := timerArg(mlr, 0)
	if err != nil {
		return err
	}

	err = t.stop()
	if err != nil {
		return err
	}

	return nil
}
