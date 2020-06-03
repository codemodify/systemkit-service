// +build windows

package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	svcMgr "golang.org/x/sys/windows/svc/mgr"

	logging "github.com/codemodify/systemkit-logging"
	spec "github.com/codemodify/systemkit-service-spec"
	"github.com/codemodify/systemkit-service/helpers"
)

var logTag = "Windows-SERVICE"

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

type windowsService struct {
	serviceSpec spec.SERVICE
}

func newServiceFromSERVICE(serviceSpec spec.SERVICE) Service {
	logging.Debugf("%s: serviceSpec object: %s", logTag, helpers.AsJSONString(serviceSpec))

	return &windowsService{
		serviceSpec: serviceSpec,
	}
}

func newServiceFromName(name string) (Service, error) {
	// quick fire
	info := newServiceFromSERVICE(spec.SERVICE{Name: name}).Info()
	if helpers.Is(info.Error, ErrServiceDoesNotExist) {
		return nil, ErrServiceDoesNotExist
	}

	// if the service exists then fetch details
	// wmic service "systemkit-test-service" get c
	serviceSpec := spec.SERVICE{
		Name:        name,
		Description: runWmicCommand("service", fmt.Sprintf("'%s'", name), "get", "Description"),
		// Documentation: "",
		Executable: runWmicCommand("service", fmt.Sprintf("'%s'", name), "get", "PathName"),
		// Args:               "",
		// WorkingDirectory:   "",
		// Environment:        "",
		// DependsOn:          "",
		// Restart:            "",
		// DelayBeforeRestart: "",
		// StdOut:             "",
		// StdErr:             "",
		// RunAsUser:          "",
		// RunAsGroup:         "",
	}

	executableWithArgs := strings.Split(serviceSpec.Executable, " ")
	if len(executableWithArgs) > 0 {
		serviceSpec.Executable = executableWithArgs[0]
		if len(executableWithArgs) > 1 {
			serviceSpec.Args = executableWithArgs[1:]
		}
	}

	return newServiceFromSERVICE(serviceSpec), nil
}

func newServiceFromPlatformTemplate(name string, template string) (Service, error) {
	return nil, ErrServiceUnsupportedRequest
}

func (thisRef *windowsService) Install() error {
	logging.Debugf("%s: attempting to install: %s", logTag, thisRef.serviceSpec.Name)

	// 1. check if service exists
	logging.Debugf("%s: check if exists: %s", logTag, thisRef.serviceSpec.Name)

	winServiceManager, winService, sError := connectAndOpenService(thisRef.serviceSpec.Name)
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

		logging.Errorf("%s: service '%s' encountered error %s", logTag, thisRef.serviceSpec.Name, sError.Error.Error())

		return sError.Error
	}

	// 2. create the system service
	logging.Debugf("%s: creating: '%s', binary: '%s', args: '%s'", logTag, thisRef.serviceSpec.Name, thisRef.serviceSpec.Executable, thisRef.serviceSpec.Args)

	var startType uint32 = svcMgr.StartAutomatic
	if !thisRef.serviceSpec.Start.AtBoot {
		startType = svcMgr.StartManual
	}

	// FIXME: revisit dependencies
	// dependencies := []string{}
	// for _, dependsOn := range thisRef.serviceSpec.DependsOn {
	// 	dependencies = append(dependencies, string(dependsOn))
	// }

	winService, err := winServiceManager.CreateService(
		thisRef.serviceSpec.Name,
		thisRef.serviceSpec.Executable,
		svcMgr.Config{
			DisplayName: thisRef.serviceSpec.Name,
			Description: thisRef.serviceSpec.Description,
			StartType:   startType,
			// ServiceStartName: thisRef.serviceSpec.Credentials.User, // FIXME:
			// Dependencies:     dependencies,
		},
		thisRef.serviceSpec.Args...,
	)
	if err != nil {
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		logging.Errorf("%s: error creating: %s, details: %v", logTag, thisRef.serviceSpec.Name, err)

		return err
	}

	winService.Close()
	winServiceManager.Disconnect()

	logging.Debugf("%s: created: '%s', binary: '%s', args: '%s'", logTag, thisRef.serviceSpec.Name, thisRef.serviceSpec.Executable, thisRef.serviceSpec.Args)

	return nil
}

func (thisRef *windowsService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s", logTag, thisRef.serviceSpec.Name)

	winServiceManager, winService, sError := connectAndOpenService(thisRef.serviceSpec.Name)
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
		logging.Errorf("%s: failed to uninstall: %s, %v", logTag, thisRef.serviceSpec.Name, err)

		return err
	}

	logging.Debugf("%s: uninstalled: %s", logTag, thisRef.serviceSpec.Name)

	return nil
}

func (thisRef *windowsService) Start() error {
	// 1.
	logging.Debugf("%s: attempting to start: %s", logTag, thisRef.serviceSpec.Name)

	winServiceManager, winService, sError := connectAndOpenService(thisRef.serviceSpec.Name)
	if sError.Type != serviceErrorSuccess {
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		if sError.Type == serviceErrorDoesNotExist {
			return ErrServiceDoesNotExist
		}

		return sError.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	// 2.
	err := winService.Start()
	if err != nil {
		if !strings.Contains(err.Error(), "already running") {
			logging.Errorf("%s: error starting: %s, %v", logTag, thisRef.serviceSpec.Name, err)

			return fmt.Errorf("error starting: %s, %v", thisRef.serviceSpec.Name, err)
		}
	}

	logging.Debugf("%s: started: %s", logTag, thisRef.serviceSpec.Name)

	return nil
}

func (thisRef *windowsService) Stop() error {
	// 1.
	logging.Debugf("%s: attempting to stop: %s", logTag, thisRef.serviceSpec.Name)

	if thisRef.serviceSpec.OnStopDelegate != nil {
		logging.Debugf("%s: OnStopDelegate before-calling: %s", logTag, thisRef.serviceSpec.Name)

		thisRef.serviceSpec.OnStopDelegate()

		logging.Debugf("%s: OnStopDelegate after-calling: %s", logTag, thisRef.serviceSpec.Name)
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

		logging.Errorf("%s: error %s, details: %s", logTag, thisRef.serviceSpec.Name, err.Error())

		return err
	}

	// 3.
	attempt := 0
	maxAttempts := 10
	wait := 3 * time.Second
	for {
		attempt++

		logging.Debugf("%s: waiting for service to stop", logTag)

		// Wait a few seconds before retrying
		time.Sleep(wait)

		// Attempt to stop the service again
		info := thisRef.Info()
		if info.Error != nil {
			if strings.Contains(info.Error.Error(), "the pipe has been ended") {
				info.IsRunning = false
			} else {
				return info.Error
			}
		}

		// If it is now running, exit the retry loop
		if !info.IsRunning {
			break
		}

		if attempt == maxAttempts {
			return errors.New("could not stop system service after multiple attempts")
		}
	}

	logging.Debugf("%s: stopped: %s", logTag, thisRef.serviceSpec.Name)

	return nil
}

func (thisRef *windowsService) Info() Info {
	result := Info{
		Error:     nil,
		Service:   thisRef.serviceSpec,
		IsRunning: false,
		PID:       -1,
	}

	// 1.
	logging.Debugf("%s: querying status: %s", logTag, thisRef.serviceSpec.Name)

	winServiceManager, winService, sError := connectAndOpenService(thisRef.serviceSpec.Name)
	if sError.Type != serviceErrorSuccess {
		if winService != nil {
			winService.Close()
		}
		if winServiceManager != nil {
			winServiceManager.Disconnect()
		}

		if sError.Type == serviceErrorDoesNotExist {
			result.Error = ErrServiceDoesNotExist
		} else {
			result.Error = sError.Error
		}

		return result
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	// 2.
	stat, err1 := winService.Query()
	if err1 != nil {
		logging.Errorf("%s: error getting service status: %s", logTag, err1)

		result.Error = fmt.Errorf("error getting service status: %v", err1)
		return result
	}

	logging.Debugf("%s: service status: %#v", logTag, stat)

	result.PID = int(stat.ProcessId)
	result.IsRunning = (stat.State == svc.Running)
	if !result.IsRunning {
		result.PID = -1
	}

	return result
}

func (thisRef *windowsService) control(serviceSpec svc.Cmd, state svc.State) error {
	logging.Debugf("%s: attempting to control: %s, cmd: %v", logTag, thisRef.serviceSpec.Name, serviceSpec)

	winServiceManager, winService, err := connectAndOpenService(thisRef.serviceSpec.Name)
	if err.Type != serviceErrorSuccess {
		return err.Error
	}
	defer winServiceManager.Disconnect()
	defer winService.Close()

	status, err1 := winService.Control(serviceSpec)
	if err1 != nil {
		logging.Errorf("%s: could not send control: %d, to: %s, details: %v", logTag, serviceSpec, thisRef.serviceSpec.Name, err1)

		return fmt.Errorf("could not send control: %d, to: %s, details: %v", serviceSpec, thisRef.serviceSpec.Name, err1)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != state {
		// Exit if a timeout is reached
		if timeout.Before(time.Now()) {
			logging.Errorf("%s: timeout waiting for service to go to state=%d", logTag, state)

			return fmt.Errorf("timeout waiting for service to go to state=%d", state)
		}

		time.Sleep(300 * time.Millisecond)

		// Make sure transition happens to the desired state
		status, err1 = winService.Query()
		if err1 != nil {
			logging.Errorf("%s: could not retrieve service status: %v", logTag, err1)

			return fmt.Errorf("could not retrieve service status: %v", err1)
		}
	}

	return nil
}

func connectAndOpenService(serviceName string) (*svcMgr.Mgr, *svcMgr.Service, serviceError) {
	// 1.
	logging.Debugf("%s: connecting to Windows Service Manager", logTag)

	winServiceManager, err := svcMgr.Connect()
	if err != nil {
		logging.Errorf("%s: error connecting to Windows Service Manager: %v", logTag, err)
		return nil, nil, serviceError{Type: serviceErrorCantConnect, Error: err}
	}

	// 2.
	logging.Debugf("%s: opening service: %s", logTag, serviceName)

	winService, err := winServiceManager.OpenService(serviceName)
	if err != nil {
		logging.Errorf("%s: error opening service: %s, %v", logTag, serviceName, err)

		return winServiceManager, nil, serviceError{Type: serviceErrorDoesNotExist, Error: err}
	}

	return winServiceManager, winService, serviceError{Type: serviceErrorSuccess}
}

func (thisRef *windowsService) Exists() bool {
	logging.Debugf("%s: checking existence: %s", logTag, thisRef.serviceSpec.Name)

	args := []string{"queryex", fmt.Sprintf("\"%s\"", thisRef.serviceSpec.Name)}

	// https://www.computerhope.com/sc-serviceSpec.htm
	logging.Debugf("%s: running: 'sc %s'", logTag, strings.Join(args, " "))

	_, err := helpers.ExecWithArgs("sc", args...)
	if err != nil {
		logging.Errorf("%s: error when checking %s", logTag, err)
		return false
	}

	return true
}

func runWmicCommand(args ...string) string {
	// wmic service "systemkit-test-service" get PathName

	logging.Debugf("%s: RUN-WMIC: wmic %s", logTag, strings.Join(args, " "))

	output, err := helpers.ExecWithArgs("wmic", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-WMIC-OUT: output: %s, error: %s", logTag, output, errAsString)

	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}

	return ""
}
