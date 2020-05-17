// +build linux

package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

var logTagUpstart = "UpStart-SERVICE"

type upstartService struct {
	serviceSpec            spec.SERVICE
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromSERVICE_Upstart(serviceSpec spec.SERVICE) Service {
	logging.Debugf("%s: serviceSpec object: %s, from %s", logTagUpstart, helpersJSON.AsJSONString(serviceSpec), helpersReflect.GetThisFuncName())

	return &upstartService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName_Upstart(name string) (Service, error) {
	serviceFile := filepath.Join("/etc/init/", name+".conf")

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromPlatformTemplate_Upstart(name, string(fileContent))
}

func newServiceFromPlatformTemplate_Upstart(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTagUpstart, template, helpersReflect.GetThisFuncName())

	serviceSpec := encoders.UpStartToSERVICE(template)

	return &upstartService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}, nil
}

func (thisRef upstartService) Install() error {
	dir := filepath.Dir(thisRef.filePath())

	// 1.
	logging.Debugf("making sure folder exists: %s, from %s", dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("generating unit file, from %s", helpersReflect.GetThisFuncName())

	fileContent := encoders.SERVICEToUpStart(thisRef.serviceSpec)

	if !thisRef.useConfigAsFileContent {
		fileContent = thisRef.fileContentTemplate
	}

	logging.Debugf("writing unit to: %s, from %s", thisRef.filePath(), helpersReflect.GetThisFuncName())

	err := ioutil.WriteFile(thisRef.filePath(), []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	logging.Debugf("wrote unit: %s, from %s", string(fileContent), helpersReflect.GetThisFuncName())

	return nil
}

func (thisRef upstartService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTagUpstart, thisRef.serviceSpec.Name, helpersReflect.GetThisFuncName())

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

func (thisRef upstartService) Start() error {
	// 1.
	logging.Debugf("loading unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err := runInitctlCommand("start", thisRef.serviceSpec.Name)
	if err != nil {
		if strings.Contains(output, "Failed to start") && strings.Contains(output, "not found") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef upstartService) Stop() error {
	// 1.
	logging.Debugf("stopping service, from %s", helpersReflect.GetThisFuncName())
	output, err := runInitctlCommand("stop", thisRef.serviceSpec.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef upstartService) Info() Info {
	fileContent, _ := ioutil.ReadFile(thisRef.filePath())

	result := Info{
		Error:       nil,
		Service:     thisRef.serviceSpec,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	// output, err := runInitctlCommand("status", thisRef.serviceSpec.Name)
	// if err != nil {
	// 	result.Error = err
	// 	return result
	// }

	// if strings.Contains(output, "could not be found") {
	// 	result.Error = ErrServiceDoesNotExist
	// 	return result
	// }

	// for _, line := range strings.Split(output, "\n") {
	// 	if strings.Contains(line, "Main PID") {
	// 		lineParts := strings.Split(strings.TrimSpace(line), " ")
	// 		if len(lineParts) >= 2 {
	// 			result.PID, _ = strconv.Atoi(lineParts[2])
	// 		}
	// 	} else if strings.Contains(line, "Active") {
	// 		if strings.Contains(line, "active (running)") {
	// 			result.IsRunning = true
	// 		}
	// 	}
	// }

	return result
}

func (thisRef upstartService) filePath() string {
	return filepath.Join("/etc/init/", thisRef.serviceSpec.Name+".conf")
}

func runInitctlCommand(args ...string) (string, error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Debugf("%s: RUN-INITCTL: initctl %s, from %s", logTagUpstart, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("initctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-INITCTL-OUT: output: %s, error: %s, from %s", logTagUpstart, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
