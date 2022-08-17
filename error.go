package udpsrv

import "fmt"

type ErrNoQueue struct{}

func (e ErrNoQueue) Error() string {
	return "no queue"
}

type ErrHaltTimedOut struct{}

func (e ErrHaltTimedOut) Error() string {
	return "halt timed out"
}

type ErrListenerAlreadyRunning struct {
	Address string
}

func (e ErrListenerAlreadyRunning) Error() string {
	return fmt.Sprintf("listener already running on %s", e.Address)
}

type ErrListenerNotRunning struct {
	Address string
}

func (e ErrListenerNotRunning) Error() string {
	return fmt.Sprintf("listener not running on %s", e.Address)
}
