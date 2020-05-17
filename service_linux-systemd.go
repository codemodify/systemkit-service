// +build linux

package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersUser "github.com/codemodify/systemkit-helpers-os"
	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
	"github.com/codemodify/systemkit-service/encoders"
	"github.com/codemodify/systemkit-service/spec"
)

var logTagSystemD = "SystemD-SERVICE"

type systemdService struct {
	serviceSpec            spec.SERVICE
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromSERVICE_SystemD(serviceSpec spec.SERVICE) Service {
	logging.Debugf("%s: spec.SERVICE object: %s, from %s", logTagSystemD, helpersJSON.AsJSONString(serviceSpec), helpersReflect.GetThisFuncName())

	return &systemdService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName_SystemD(name string) (Service, error) {
	fileContent := []byte{}
	var err error

	if helpersUser.IsRoot() {
		serviceFile := filepath.Join("/etc/systemd/system", name+".service")
		fileContent, err = ioutil.ReadFile(serviceFile)
		if err != nil {
			serviceFile = filepath.Join("/usr/lib/systemd/system", name+".service")
			fileContent, err = ioutil.ReadFile(serviceFile)
			if err != nil {
				return nil, ErrServiceDoesNotExist
			}
		}
	} else {
		serviceFile := filepath.Join(helpersUser.HomeDir(""), ".config/systemd/user", name+".service")
		fileContent, err = ioutil.ReadFile(serviceFile)
		if err != nil {
			return nil, ErrServiceDoesNotExist
		}
	}

	return newServiceFromPlatformTemplate_SystemD(name, string(fileContent))
}

func newServiceFromPlatformTemplate_SystemD(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTagSystemD, template, helpersReflect.GetThisFuncName())

	serviceSpec := encoders.SystemDToSERVICE(template)

	return &systemdService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}, nil
}

func (thisRef systemdService) Install() error {
	dir := filepath.Dir(thisRef.filePath())

	// 1.
	logging.Debugf("making sure folder exists: %s, from %s", dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("generating unit file, from %s", helpersReflect.GetThisFuncName())

	fileContent := encoders.SERVICEToSystemD(thisRef.serviceSpec)

	if !thisRef.useConfigAsFileContent {
		fileContent = thisRef.fileContentTemplate
	}

	logging.Debugf("writing unit to: %s, from %s", thisRef.filePath(), helpersReflect.GetThisFuncName())

	err := ioutil.WriteFile(thisRef.filePath(), []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	logging.Debugf("wrote unit: %s, from %s", fileContent, helpersReflect.GetThisFuncName())

	return nil
}

func (thisRef systemdService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTagSystemD, thisRef.serviceSpec.Name, helpersReflect.GetThisFuncName())

	// 2.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 3.
	logging.Debugf("remove unit file, from %s", helpersReflect.GetThisFuncName())
	err = os.Remove(thisRef.filePath())
	if e, ok := err.(*os.PathError); ok {
		if os.IsNotExist(e.Err) {
			return nil
		}
	}

	return err
}

func (thisRef systemdService) Start() error {
	// 1.
	logging.Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	output, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Debugf("enabling unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("enable", thisRef.serviceSpec.Name)
	if err != nil {
		if strings.Contains(output, "Failed to enable unit") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Debugf("loading unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("start", thisRef.serviceSpec.Name)
	if err != nil {
		if strings.Contains(output, "Failed to start") && strings.Contains(output, "not found") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef systemdService) Stop() error {
	// 1.
	logging.Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	_, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Debugf("stopping unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err := runSystemCtlCommand("stop", thisRef.serviceSpec.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Debugf("disabling unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("disable", thisRef.serviceSpec.Name)
	if err != nil {
		logging.Warningf("stopping unit file with systemd, from %s", helpersReflect.GetThisFuncName())

		if strings.Contains(output, "Failed to disable") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		} else if strings.Contains(output, "Removed") {
			return nil
		}

		return err
	}

	// 4.
	logging.Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	_, err = runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 5.
	logging.Debugf("running reset-failed, from %s", helpersReflect.GetThisFuncName())
	_, err = runSystemCtlCommand("reset-failed")
	if err != nil {
		return err
	}

	return nil
}

func (thisRef systemdService) Info() Info {
	fileContent, _ := ioutil.ReadFile(thisRef.filePath())

	result := Info{
		Error:       nil,
		Service:     thisRef.serviceSpec,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	output, err := runSystemCtlCommand("status", thisRef.serviceSpec.Name)
	if err != nil {
		result.Error = err
		return result
	}

	if strings.Contains(output, "could not be found") {
		result.Error = ErrServiceDoesNotExist
		return result
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Main PID") {
			lineParts := strings.Split(strings.TrimSpace(line), " ")
			if len(lineParts) >= 2 {
				result.PID, _ = strconv.Atoi(lineParts[2])
			}
		} else if strings.Contains(line, "Active") {
			if strings.Contains(line, "active (running)") {
				result.IsRunning = true
			}
		}
	}

	return result
}

func (thisRef systemdService) filePath() string {
	if helpersUser.IsRoot() {
		return filepath.Join("/etc/systemd/system", thisRef.serviceSpec.Name+".service")
	}

	return filepath.Join(helpersUser.HomeDir(""), ".config/systemd/user", thisRef.serviceSpec.Name+".service")
}

func runSystemCtlCommand(args ...string) (string, error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Debugf("%s: RUN-SYSTEMCTL: systemctl %s, from %s", logTagSystemD, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("systemctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-SYSTEMCTL-OUT: output: %s, error: %s, from %s", logTagSystemD, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
