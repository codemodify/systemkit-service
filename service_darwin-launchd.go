// +build darwin

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

var logTag = "LaunchD-SERVICE"

type launchdService struct {
	serviceSpec            spec.SERVICE
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromSERVICE(serviceSpec spec.SERVICE) Service {
	// override some values - platform specific
	// https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
	logDir := filepath.Join(helpersUser.HomeDir(""), "Library/Logs", serviceSpec.Name)
	if helpersUser.IsRoot() {
		logDir = filepath.Join("/Library/Logs", serviceSpec.Name)
	}

	if serviceSpec.Logging.StdOut.UseDefault {
		serviceSpec.Logging.StdOut.Value = filepath.Join(logDir, serviceSpec.Name+".stdout.log")
	}

	if serviceSpec.Logging.StdErr.UseDefault {
		serviceSpec.Logging.StdErr.Value = filepath.Join(logDir, serviceSpec.Name+".stderr.log")
	}

	logging.Debugf("%s: serviceSpec object: %s, from %s", logTag, helpersJSON.AsJSONString(serviceSpec), helpersReflect.GetThisFuncName())

	launchdService := &launchdService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: true,
	}

	return launchdService
}

func newServiceFromName(name string) (Service, error) {
	serviceFile := filepath.Join(helpersUser.HomeDir(""), "Library/LaunchAgents", name+".plist")
	if helpersUser.IsRoot() {
		serviceFile = filepath.Join("/Library/LaunchDaemons", name+".plist")
	}

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromPlatformTemplate(name, string(fileContent))
}

func newServiceFromPlatformTemplate(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTag, template, helpersReflect.GetThisFuncName())

	return &launchdService{
		serviceSpec:            encoders.LaunchDToSERVICE(template),
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}, nil
}

func (thisRef launchdService) Install() error {
	dir := filepath.Dir(thisRef.filePath())

	// 1.
	logging.Debugf("%s: making sure folder exists: %s, from %s", logTag, dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("%s: generating plist file, from %s", logTag, helpersReflect.GetThisFuncName())
	fileContent, err := encoders.SERVICEToLaunchD(thisRef.serviceSpec)
	if err != nil {
		return err
	}

	logging.Debugf("%s: writing plist to: %s, from %s", logTag, thisRef.filePath(), helpersReflect.GetThisFuncName())
	err = ioutil.WriteFile(thisRef.filePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Debugf("%s: wrote unit: %s, from %s", logTag, string(fileContent), helpersReflect.GetThisFuncName())

	return nil
}

func (thisRef launchdService) Uninstall() error {
	// 1.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 2.
	logging.Debugf("%s: remove plist file: %s, from %s", logTag, thisRef.filePath(), helpersReflect.GetThisFuncName())
	err = os.Remove(thisRef.filePath())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such file or directory") {
			return nil
		}

		return err
	}

	// INFO: ignore the return value as is it is barely defined by the docs
	// what the expected behavior would be. The previous stop and remove the "plist" file
	// will uninstall the service.
	runLaunchCtlCommand("remove", thisRef.serviceSpec.Name)
	return nil
}

func (thisRef launchdService) Start() error {
	// 1.
	output, _ := runLaunchCtlCommand("load", "-w", thisRef.filePath())
	if strings.Contains(output, "No such file or directory") {
		return ErrServiceDoesNotExist
	} else if strings.Contains(output, "Invalid property list") {
		return ErrServiceConfigError
	}

	if strings.Contains(output, "service already loaded") {
		logging.Debugf("service already loaded, from %s", helpersReflect.GetThisFuncName())

		return nil
	}

	runLaunchCtlCommand("start", thisRef.serviceSpec.Name)
	return nil
}

func (thisRef launchdService) Stop() error {
	runLaunchCtlCommand("stop", thisRef.serviceSpec.Name)
	output, err := runLaunchCtlCommand("unload", thisRef.filePath())
	if strings.Contains(output, "Could not find specified service") {
		return ErrServiceDoesNotExist
	}

	return err
}

func (thisRef launchdService) Info() Info {
	fileContent, fileContentErr := ioutil.ReadFile(thisRef.filePath())

	result := Info{
		Error:       nil,
		Service:     thisRef.serviceSpec,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	if fileContentErr != nil || len(fileContent) <= 0 {
		result.Error = ErrServiceDoesNotExist
	}

	output, err := runLaunchCtlCommand("list")
	if err != nil {
		result.Error = err
		logging.Errorf("error getting launchctl status: %s, from %s", err, helpersReflect.GetThisFuncName())
		return result
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		chunks := strings.Split(line, "\t")

		if chunks[2] == thisRef.serviceSpec.Name {
			if chunks[0] != "-" {
				pid, _ := strconv.Atoi(chunks[0])
				result.PID = pid
			}

			if result.PID != -1 {
				result.IsRunning = true
			}

			break
		}
	}

	return result
}

func (thisRef launchdService) filePath() string {
	if helpersUser.IsRoot() {
		return filepath.Join("/Library/LaunchDaemons", thisRef.serviceSpec.Name+".plist")
	}

	return filepath.Join(helpersUser.HomeDir(""), "Library/LaunchAgents", thisRef.serviceSpec.Name+".plist")
}

func runLaunchCtlCommand(args ...string) (string, error) {
	// if !helpersUser.IsRoot() {
	// 	args = append([]string{"--user"}, args...)
	// }

	logging.Debugf("%s: RUN-LAUNCHCTL: launchctl %s, from %s", logTag, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("launchctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-LAUNCHCTL-OUT: output: %s, error: %s, from %s", logTag, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
