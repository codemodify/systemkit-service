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
	logging "github.com/codemodify/systemkit-logging"
	encoders "github.com/codemodify/systemkit-service-encoders-systemv"
	spec "github.com/codemodify/systemkit-service-spec"
)

var logTagSystemV = "SystemV-SERVICE"

type systemvService struct {
	serviceSpec            spec.SERVICE
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromSERVICE_SystemV(serviceSpec spec.SERVICE) Service {
	logging.Debugf("%s: serviceSpec object: %s", logTagSystemV, helpersJSON.AsJSONString(serviceSpec))

	return &systemvService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName_SystemV(name string) (Service, error) {
	serviceFile := filepath.Join("/etc/init.d/", name)

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromPlatformTemplate_SystemV(name, string(fileContent))
}

func newServiceFromPlatformTemplate_SystemV(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s", logTagSystemV, template)

	serviceSpec := encoders.SystemVToSERVICE(template)

	return &systemvService{
		serviceSpec:            serviceSpec,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}, nil
}

func (thisRef systemvService) Install() error {
	dir := filepath.Dir(thisRef.filePath())

	// 1.
	logging.Debugf("making sure folder exists: %s", dir)
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("generating unit file")

	fileContent := encoders.SERVICEToSystemV(thisRef.serviceSpec)

	if !thisRef.useConfigAsFileContent {
		fileContent = thisRef.fileContentTemplate
	}

	logging.Debugf("writing unit to: %s", thisRef.filePath())

	err := ioutil.WriteFile(thisRef.filePath(), []byte(fileContent), 0755)
	if err != nil {
		return err
	}

	// additional rc.d magic
	for _, i := range [...]string{"2", "3", "4", "5"} {
		if err = os.Symlink(thisRef.filePath(), "/etc/rc"+i+".d/S50"+thisRef.serviceSpec.Name); err != nil {
			continue
		}
	}
	for _, i := range [...]string{"0", "1", "6"} {
		if err = os.Symlink(thisRef.filePath(), "/etc/rc"+i+".d/K02"+thisRef.serviceSpec.Name); err != nil {
			continue
		}
	}

	logging.Debugf("wrote unit: %s", fileContent)

	return nil
}

func (thisRef systemvService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s", logTagSystemV, thisRef.serviceSpec.Name)

	// 2.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 3.
	logging.Debugf("remove unit file")
	err = os.Remove(thisRef.filePath())
	if e, ok := err.(*os.PathError); ok {
		if os.IsNotExist(e.Err) {
			return nil
		}
	}

	return err
}

func (thisRef systemvService) Start() error {
	// 1.
	logging.Debugf("loading unit file with systemd")
	output, err := runServiceCommand(thisRef.serviceSpec.Name, "start")
	if err != nil {
		if strings.Contains(output, "Failed to start") && strings.Contains(output, "not found") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef systemvService) Stop() error {
	// 1.
	logging.Debugf("stopping service")
	output, err := runServiceCommand(thisRef.serviceSpec.Name, "stop")
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef systemvService) Info() Info {
	fileContent, _ := ioutil.ReadFile(thisRef.filePath())

	result := Info{
		Error:       nil,
		Service:     thisRef.serviceSpec,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	// output, err := runServiceCommand("status", thisRef.serviceSpec.Name)
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

func (thisRef systemvService) filePath() string {
	return filepath.Join("/etc/init.d/", thisRef.serviceSpec.Name)
}

func runServiceCommand(args ...string) (string, error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Debugf("%s: RUN-SERVICE: service %s", logTagSystemV, strings.Join(args, " "))

	output, err := helpersExec.ExecWithArgs("service", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-SERVICE-OUT: output: %s, error: %s", logTagSystemV, output, errAsString)

	return output, err
}
