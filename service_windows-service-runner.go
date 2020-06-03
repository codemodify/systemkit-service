// +build windows

package service

import (
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"

	logging "github.com/codemodify/systemkit-logging"
	spec "github.com/codemodify/systemkit-service-spec"
)

// WindowsServiceQueryHandler -
type WindowsServiceQueryHandler interface {
	RunServiceQueryLoop() error
}

type windowsServiceQueryHandler struct {
	serviceSpec *spec.SERVICE
}

// NewWindowsServiceQueryHandler -
func NewWindowsServiceQueryHandler(serviceSpec *spec.SERVICE) WindowsServiceQueryHandler {
	return &windowsServiceQueryHandler{
		serviceSpec: serviceSpec,
	}
}

func (thisRef windowsServiceQueryHandler) RunServiceQueryLoop() error {
	logging.Debugf("%s: attempting to run: %s", logTag, thisRef.serviceSpec.Name)

	wg := sync.WaitGroup{}

	wg.Add(1)
	var err error
	go func() {
		err = svc.Run(thisRef.serviceSpec.Name, thisRef)
		wg.Done()
	}()

	logging.Debugf("%s: running: %s", logTag, thisRef.serviceSpec.Name)
	wg.Wait()

	if err != nil {
		logging.Errorf("%s: failed to run: %s, %v", logTag, thisRef.serviceSpec.Name, err)
	}

	logging.Debugf("%s: stopped: %s", logTag, thisRef.serviceSpec.Name)

	return nil
}

func (thisRef windowsServiceQueryHandler) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	logging.Debugf("%s: WINDOWS SERVICE EXECUTE", logTag)

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus

			case svc.Stop, svc.Shutdown:
				if thisRef.serviceSpec.OnStopDelegate != nil {
					logging.Debugf("%s: OnStopDelegate before-calling: %s", logTag, thisRef.serviceSpec.Name)

					go thisRef.serviceSpec.OnStopDelegate()

					logging.Debugf("%s: OnStopDelegate after-calling: %s", logTag, thisRef.serviceSpec.Name)
				}

				break loop

			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}

			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

			default:
				logging.Warningf("%s: unexpected control request #%d", logTag, c)
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}
