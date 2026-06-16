package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	rt "github.com/arnodel/golua/runtime"
)

type timerType int64

const (
	timerInterval timerType = iota
	timerTimeout
)

// #type
// #interface timers
// #property type What type of timer it is
// #property running If the timer is running
// #property duration The duration in milliseconds that the timer will run
// The Job type describes a Hilbish timer.
type timer struct {
	mu      sync.Mutex
	id      int
	typ     timerType
	running bool
	dur     time.Duration
	fun     *rt.Closure
	th      *timersModule
	ticker  *time.Ticker
	ud      *rt.UserData
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
				_, err := rt.Call1(l.MainThread(), rt.FunctionValue(t.fun))
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

// #interface timers
// #member
// start()
// Starts a timer.
func timerStart(thr *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	t, err := timerArg(c, 0)
	if err != nil {
		return nil, err
	}

	err = t.start()
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}

// #interface timers
// #member
// stop()
// Stops a timer.
func timerStop(thr *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	t, err := timerArg(c, 0)
	if err != nil {
		return nil, err
	}

	err = t.stop()
	if err != nil {
		return nil, err
	}

	return c.Next(), nil
}
