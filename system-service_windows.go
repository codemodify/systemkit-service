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

	helpersExec "github.com/codemodify/systemkit-helpers"
	helpersJSON "github.com/codemodify/systemkit-helpers"
	helpersReflect "github.com/codemodify/systemkit-helpers"
	logging "github.com/codemodify/systemkit-logging"
	loggingC "github.com/codemodify/systemkit-logging/contracts"
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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: config object: %s ", logTag, helpersJSON.AsJSONString(command)),
	})

	return &WindowsService{
		command: command,
	}
}

// Run -
func (thisRef *WindowsService) Run() error {
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to run: %s", logTag, thisRef.command.Name),
	})

	wg := sync.WaitGroup{}

	wg.Add(1)
	var err error
	go func() {
		err = svc.Run(thisRef.command.Name, thisRef)
		wg.Done()
	}()

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: running: %s", logTag, thisRef.command.Name),
	})
	wg.Wait()

	if err != nil {
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: failed to run: %s, %v", logTag, thisRef.command.Name, err),
		})
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: stopped: %s", logTag, thisRef.command.Name),
	})

	return nil
}

// Install -
func (thisRef *WindowsService) Install(start bool) error {
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to install: %s", logTag, thisRef.command.Name),
	})

	// 1. check if service exists
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: check if exists: %s", logTag, thisRef.command.Name),
	})

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

		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: service '%s' encountered error %s", logTag, thisRef.command.Name, sError.Error.Error()),
		})

		return sError.Error
	}

	// 2. create the system service
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: creating: '%s', binary: '%s', args: '%s'", logTag, thisRef.command.Name, thisRef.command.Executable, thisRef.command.Args),
	})

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

		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: error creating: %s, details: %v ", logTag, thisRef.command.Name, err),
		})

		return err
	}

	winService.Close()
	winServiceManager.Disconnect()

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: created: '%s', binary: '%s', args: '%s'", logTag, thisRef.command.Name, thisRef.command.Executable, thisRef.command.Args),
	})

	// 3. start if needed
	if start {
		return thisRef.Start()
	}

	return nil
}

// Start -
func (thisRef *WindowsService) Start() error {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to start: %s", logTag, thisRef.command.Name),
	})

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
			logging.Instance().LogErrorWithFields(loggingC.Fields{
				"method":  helpersReflect.GetThisFuncName(),
				"message": fmt.Sprintf("%s: error starting: %s, %v", logTag, thisRef.command.Name, err),
			})

			return fmt.Errorf("error starting: %s, %v", thisRef.command.Name, err)
		}
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: started: %s", logTag, thisRef.command.Name),
	})

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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to stop: %s", logTag, thisRef.command.Name),
	})

	if thisRef.command.OnStopDelegate != nil {
		logging.Instance().LogDebugWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: OnStopDelegate before-calling: %s", logTag, thisRef.command.Name),
		})

		thisRef.command.OnStopDelegate()

		logging.Instance().LogDebugWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: OnStopDelegate after-calling: %s", logTag, thisRef.command.Name),
		})
	}

	// 2.
	err := thisRef.control(svc.Stop, svc.Stopped)
	if err != nil {
		e := err.Error()
		if strings.Contains(e, "service does not exist") {
			return ErrServiceDoesNotExist
		} else if strings.Contains(e, "service has not been started") {
			return nil
		}

		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: error %s, details: %s", logTag, thisRef.command.Name, err.Error()),
		})

		return err
	}

	// 3.
	attempt := 0
	maxAttempts := 10
	wait := 3 * time.Second
	for {
		attempt++

		logging.Instance().LogDebugWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: waiting for service to stop", logTag),
		})

		// Wait a few seconds before retrying
		time.Sleep(wait)

		// Attempt to start the service again
		stat := thisRef.Status()
		if stat.Error != nil {
			return err
		}

		// If it is now running, exit the retry loop
		if !stat.IsRunning {
			break
		}

		if attempt == maxAttempts {
			return errors.New("could not stop system service after multiple attempts")
		}
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: stopped: %s", logTag, thisRef.command.Name),
	})

	return nil
}

// Uninstall -
func (thisRef *WindowsService) Uninstall() error {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to uninstall: %s", logTag, thisRef.command.Name),
	})

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
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: failed to uninstall: %s, %v", logTag, thisRef.command.Name, err),
		})

		return err
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: uninstalled: %s", logTag, thisRef.command.Name),
	})

	return nil
}

// Status -
func (thisRef *WindowsService) Status() Status {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: querying status: %s", logTag, thisRef.command.Name),
	})

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
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprint("%s: error getting service status: %s", logTag, err1),
		})

		return Status{
			Error: fmt.Errorf("error getting service status: %v", err1),
		}
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: service status: %#v", logTag, stat),
	})

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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: checking existance: %s", logTag, thisRef.command.Name),
	})

	args := []string{"queryex", fmt.Sprintf("\"%s\"", thisRef.command.Name)}

	// https://www.computerhope.com/sc-command.htm
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: running: 'sc %s'", logTag, strings.Join(args, " ")),
	})

	_, err := helpersExec.ExecWithArgs("sc", args...)
	if err != nil {
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: error when checking %s: ", logTag, err),
		})

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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: WINDOWS SERVICE EXECUTE", logTag),
	})

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
					logging.Instance().LogDebugWithFields(loggingC.Fields{
						"method":  helpersReflect.GetThisFuncName(),
						"message": fmt.Sprintf("%s: OnStopDelegate before-calling: %s", logTag, thisRef.command.Name),
					})

					go thisRef.command.OnStopDelegate()

					logging.Instance().LogDebugWithFields(loggingC.Fields{
						"method":  helpersReflect.GetThisFuncName(),
						"message": fmt.Sprintf("%s: OnStopDelegate after-calling: %s", logTag, thisRef.command.Name),
					})
				}

				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				// testOutput := strings.Join(args, "-")
				// testOutput += fmt.Sprintf("-%d", c.Context)
				// logging.Instance().LogDebugWithFields(loggingC.Fields{
				// 	"method":  helpersReflect.GetThisFuncName(),
				// 	"message": fmt.Sprintf("%s: %", logTag, testOutput),
				// })

				break loop

			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}

			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

			default:
				logging.Instance().LogWarningWithFields(loggingC.Fields{
					"method":  helpersReflect.GetThisFuncName(),
					"message": fmt.Sprintf("%s: unexpected control request #%d", logTag, c),
				})
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}

func (thisRef *WindowsService) control(command svc.Cmd, state svc.State) error {
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to control: %s, cmd: %v", logTag, thisRef.command.Name, command),
	})

	winServiceManager, winService, err := connectAndOpenService(thisRef.command.Name)
	if err.Type != serviceErrorSuccess {
		return err.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	status, err1 := winService.Control(command)
	if err1 != nil {
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: could not send control: %d, to: %s, details: %v", logTag, command, thisRef.command.Name, err1),
		})

		return fmt.Errorf("could not send control: %d, to: %s, details: %v", command, thisRef.command.Name, err1)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != state {
		// Exit if a timeout is reached
		if timeout.Before(time.Now()) {
			logging.Instance().LogErrorWithFields(loggingC.Fields{
				"method":  helpersReflect.GetThisFuncName(),
				"message": fmt.Sprintf("%s: timeout waiting for service to go to state=%d", logTag, state),
			})

			return fmt.Errorf("timeout waiting for service to go to state=%d", state)
		}

		time.Sleep(300 * time.Millisecond)

		// Make sure transition happens to the desired state
		status, err1 = winService.Query()
		if err1 != nil {
			logging.Instance().LogErrorWithFields(loggingC.Fields{
				"method":  helpersReflect.GetThisFuncName(),
				"message": fmt.Sprintf("%s: could not retrieve service status: %v", logTag, err1),
			})

			return fmt.Errorf("could not retrieve service status: %v", err1)
		}
	}

	return nil
}

func connectAndOpenService(serviceName string) (*svcMgr.Mgr, *svcMgr.Service, serviceError) {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: connecting to Windows Service Manager", logTag),
	})

	winServiceManager, err := svcMgr.Connect()
	if err != nil {
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: error connecting to Windows Service Manager: %v", logTag, err),
		})
		return nil, nil, serviceError{Type: serviceErrorCantConnect, Error: err}
	}

	// 2.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: opening service: %s", logTag, serviceName),
	})

	winService, err := winServiceManager.OpenService(serviceName)
	if err != nil {
		logging.Instance().LogErrorWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": fmt.Sprintf("%s: error opening service: %s, %v", logTag, serviceName, err),
		})

		return winServiceManager, nil, serviceError{Type: serviceErrorDoesNotExist, Error: err}
	}

	return winServiceManager, winService, serviceError{Type: serviceErrorSuccess}
}
