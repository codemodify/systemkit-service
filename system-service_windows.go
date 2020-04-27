// +build windows

package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"
	svcMgr "golang.org/x/sys/windows/svc/mgr"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

var logTag = "WINDOWS-SERVICE"

type serviceErrorType int

const (
	serviceErrorSuccess      serviceErrorType = iota
	serviceErrorDoesNotExist                  = iota
	serviceErrorCantConnect                   = iota
	serviceErrorOther                         = iota
)

func (thisRef serviceErrorType) String() string {
	switch thisRef {
	case serviceErrorSuccess:
		return "Success"

	case serviceErrorDoesNotExist:
		return "Service Does Not Exist"

	case serviceErrorCantConnect:
		return "Service Can't Connect"

	case serviceErrorOther:
		return "Other error occured"

	default:
		return fmt.Sprintf("%d", int(thisRef))
	}
}

type serviceError struct {
	Type  serviceErrorType
	Error error
}

// WindowsService - Represents Windows service
type WindowsService struct {
	command Command
}

// New -
func New(command Command) SystemService {
	logging.Debugf("%s: config object: %s, from %s", logTag, helpersJSON.AsJSONString(command), helpersReflect.GetThisFuncName())

	return &WindowsService{
		command: command,
	}
}

// Run -
func (thisRef *WindowsService) Run() error {
	logging.Debugf("%s: attempting to run: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	wg := sync.WaitGroup{}

	wg.Add(1)
	var err error
	go func() {
		err = svc.Run(thisRef.command.Name, thisRef)
		wg.Done()
	}()

	logging.Debugf("%s: running: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())
	wg.Wait()

	if err != nil {
		logging.Errorf("%s: failed to run: %s, %v, from %s", logTag, thisRef.command.Name, err, helpersReflect.GetThisFuncName())
	}

	logging.Debugf("%s: stopped: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	return nil
}

// Install -
func (thisRef *WindowsService) Install(start bool) error {
	logging.Debugf("%s: attempting to install: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	// 1. check if service exists
	logging.Debugf("%s: check if exists: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	winServiceManager, winService, sError := connectAndOpenService(thisRef.command.Name)
	if sError.Type == serviceErrorSuccess { // service already exists
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		return nil
	}

	if sError.Type != serviceErrorDoesNotExist { // if any other error then return it
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		logging.Errorf("%s: service '%s' encountered error %s, from %s", logTag, thisRef.command.Name, sError.Error.Error(), helpersReflect.GetThisFuncName())

		return sError.Error
	}

	// 2. create the system service
	logging.Debugf("%s: creating: '%s', binary: '%s', args: '%s', from %s", logTag, thisRef.command.Name, thisRef.command.Executable, thisRef.command.Args, helpersReflect.GetThisFuncName())

	winService, err := winServiceManager.CreateService(
		thisRef.command.Name,
		thisRef.command.Executable,
		svcMgr.Config{
			StartType:   svcMgr.StartAutomatic,
			DisplayName: thisRef.command.Name,
			Description: thisRef.command.Description,
		},
		thisRef.command.Args...,
	)
	if err != nil {
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		logging.Errorf("%s: error creating: %s, details: %v, from %s", logTag, thisRef.command.Name, err, helpersReflect.GetThisFuncName())

		return err
	}

	winService.Close()
	winServiceManager.Disconnect()

	logging.Debugf("%s: created: '%s', binary: '%s', args: '%s', from %s", logTag, thisRef.command.Name, thisRef.command.Executable, thisRef.command.Args, helpersReflect.GetThisFuncName())

	// 3. start if needed
	if start {
		return thisRef.Start()
	}

	return nil
}

// Start -
func (thisRef *WindowsService) Start() error {
	// 1.
	logging.Debugf("%s: attempting to start: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	winServiceManager, winService, sError := connectAndOpenService(thisRef.command.Name)
	if sError.Type != serviceErrorSuccess {
		return sError.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	// 2.
	err := winService.Start()
	if err != nil {
		if !strings.Contains(err.Error(), "already running") {
			logging.Errorf("%s: error starting: %s, %v, from %s", logTag, thisRef.command.Name, err, helpersReflect.GetThisFuncName())

			return fmt.Errorf("error starting: %s, %v", thisRef.command.Name, err)
		}
	}

	logging.Debugf("%s: started: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	return nil
}

// Restart -
func (thisRef *WindowsService) Restart() error {
	if err := thisRef.Stop(); err != nil {
		return err
	}

	return thisRef.Start()
}

// Stop -
func (thisRef *WindowsService) Stop() error {
	// 1.
	logging.Debugf("%s: attempting to stop: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	if thisRef.command.OnStopDelegate != nil {
		logging.Debugf("%s: OnStopDelegate before-calling: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

		thisRef.command.OnStopDelegate()

		logging.Debugf("%s: OnStopDelegate after-calling: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())
	}

	// 2.
	err := thisRef.control(svc.Stop, svc.Stopped)
	if err != nil {
		e := err.Error()
		if strings.Contains(e, "service does not exist") {
			return ErrServiceDoesNotExist
		} else if strings.Contains(e, "service has not been started") {
			return nil
		} else if strings.Contains(e, "the pipe has been ended") {
			return nil
		}

		logging.Errorf("%s: error %s, details: %s, from %s", logTag, thisRef.command.Name, err.Error(), helpersReflect.GetThisFuncName())

		return err
	}

	// 3.
	attempt := 0
	maxAttempts := 10
	wait := 3 * time.Second
	for {
		attempt++

		logging.Debugf("%s: waiting for service to stop, from %s", logTag, helpersReflect.GetThisFuncName())

		// Wait a few seconds before retrying
		time.Sleep(wait)

		// Attempt to stop the service again
		stat := thisRef.Status()
		if stat.Error != nil {
			if strings.Contains(stat.Error(), "the pipe has been ended") {
				stat.IsRunning = false
			} else {
				return stat.Error
			}
		}

		// If it is now running, exit the retry loop
		if !stat.IsRunning {
			break
		}

		if attempt == maxAttempts {
			return errors.New("could not stop system service after multiple attempts")
		}
	}

	logging.Debugf("%s: stopped: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	return nil
}

// Uninstall -
func (thisRef *WindowsService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	winServiceManager, winService, sError := connectAndOpenService(thisRef.command.Name)
	if sError.Type == serviceErrorDoesNotExist {
		return nil
	} else if sError.Type != serviceErrorSuccess {
		return sError.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	// 2.
	err := winService.Delete()
	if err != nil {
		logging.Errorf("%s: failed to uninstall: %s, %v, from %s", logTag, thisRef.command.Name, err, helpersReflect.GetThisFuncName())

		return err
	}

	logging.Debugf("%s: uninstalled: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	return nil
}

// Status -
func (thisRef *WindowsService) Status() Status {
	// 1.
	logging.Debugf("%s: querying status: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	winServiceManager, winService, err := connectAndOpenService(thisRef.command.Name)
	if err.Type != serviceErrorSuccess {
		return Status{
			Error: err.Error,
		}
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	// 2.
	stat, err1 := winService.Query()
	if err1 != nil {
		logging.Errorf("%s: error getting service status: %s, from %s", logTag, err1, helpersReflect.GetThisFuncName())

		return Status{
			Error: fmt.Errorf("error getting service status: %v", err1),
		}
	}

	logging.Debugf("%s: service status: %#v, from %s", logTag, stat, helpersReflect.GetThisFuncName())

	status := Status{
		PID:       int(stat.ProcessId),
		IsRunning: stat.State == svc.Running,
	}
	if !status.IsRunning {
		status.PID = -1
	}

	return status
}

// Exists -
func (thisRef *WindowsService) Exists() bool {
	logging.Debugf("%s: checking existence: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	args := []string{"queryex", fmt.Sprintf("\"%s\"", thisRef.command.Name)}

	// https://www.computerhope.com/sc-command.htm
	logging.Debugf("%s: running: 'sc %s', from %s", logTag, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	_, err := helpersExec.ExecWithArgs("sc", args...)
	if err != nil {
		logging.Errorf("%s: error when checking %s, from %s", logTag, err, helpersReflect.GetThisFuncName())
		return false
	}

	return true
}

// FilePath -
func (thisRef *WindowsService) FilePath() string {
	return ""
}

// FileContent -
func (thisRef *WindowsService) FileContent() ([]byte, error) {
	return []byte{}, nil
}

// Execute - implement the Windows `service.Handler` interface
func (thisRef *WindowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	logging.Debugf("%s: WINDOWS SERVICE EXECUTE, from %s", logTag, helpersReflect.GetThisFuncName())

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
				if thisRef.command.OnStopDelegate != nil {
					logging.Debugf("%s: OnStopDelegate before-calling: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

					go thisRef.command.OnStopDelegate()

					logging.Debugf("%s: OnStopDelegate after-calling: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())
				}

				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				// testOutput := strings.Join(args, "-")
				// testOutput += fmt.Sprintf("-%d", c.Context)
				// logging.LogDebugWithFields(logging.Fields{
				// 	"method":  helpersReflect.GetThisFuncName(),
				// 	"message": fmt.Sprintf("%s: %", logTag, testOutput),
				// })

				break loop

			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}

			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

			default:
				logging.Warningf("%s: unexpected control request #%d, from %s", logTag, c, helpersReflect.GetThisFuncName())
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}

func (thisRef *WindowsService) control(command svc.Cmd, state svc.State) error {
	logging.Debugf("%s: attempting to control: %s, cmd: %v, from %s", logTag, thisRef.command.Name, command, helpersReflect.GetThisFuncName())

	winServiceManager, winService, err := connectAndOpenService(thisRef.command.Name)
	if err.Type != serviceErrorSuccess {
		return err.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	status, err1 := winService.Control(command)
	if err1 != nil {
		logging.Errorf("%s: could not send control: %d, to: %s, details: %v, from %s", logTag, command, thisRef.command.Name, err1, helpersReflect.GetThisFuncName())

		return fmt.Errorf("could not send control: %d, to: %s, details: %v", command, thisRef.command.Name, err1)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != state {
		// Exit if a timeout is reached
		if timeout.Before(time.Now()) {
			logging.Errorf("%s: timeout waiting for service to go to state=%d, from %s", logTag, state, helpersReflect.GetThisFuncName())

			return fmt.Errorf("timeout waiting for service to go to state=%d", state)
		}

		time.Sleep(300 * time.Millisecond)

		// Make sure transition happens to the desired state
		status, err1 = winService.Query()
		if err1 != nil {
			logging.Errorf("%s: could not retrieve service status: %v, from %s", logTag, err1, helpersReflect.GetThisFuncName())

			return fmt.Errorf("could not retrieve service status: %v", err1)
		}
	}

	return nil
}

func connectAndOpenService(serviceName string) (*svcMgr.Mgr, *svcMgr.Service, serviceError) {
	// 1.
	logging.Debugf("%s: connecting to Windows Service Manager, from %s", logTag, helpersReflect.GetThisFuncName())

	winServiceManager, err := svcMgr.Connect()
	if err != nil {
		logging.Errorf("%s: error connecting to Windows Service Manager: %v, from %s", logTag, err, helpersReflect.GetThisFuncName())
		return nil, nil, serviceError{Type: serviceErrorCantConnect, Error: err}
	}

	// 2.
	logging.Debugf("%s: opening service: %s, from %s", logTag, serviceName, helpersReflect.GetThisFuncName())

	winService, err := winServiceManager.OpenService(serviceName)
	if err != nil {
		logging.Errorf("%s: error opening service: %s, %v, from %s", logTag, serviceName, err, helpersReflect.GetThisFuncName())

		return winServiceManager, nil, serviceError{Type: serviceErrorDoesNotExist, Error: err}
	}

	return winServiceManager, winService, serviceError{Type: serviceErrorSuccess}
}
