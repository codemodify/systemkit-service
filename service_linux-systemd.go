// +build linux

package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersUser "github.com/codemodify/systemkit-helpers-os"
	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

var logTag = "SYSTEMD-SERVICE"

type systemdService struct {
	config                 Config
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromConfig(config Config) Service {

	config.DependsOn = append(config.DependsOn, "network.target")

	logging.Debugf("%s: config object: %s, from %s", logTag, helpersJSON.AsJSONString(config), helpersReflect.GetThisFuncName())

	return &systemdService{
		config:                 config,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName(name string) (Service, error) {
	serviceFile := filepath.Join(helpersUser.HomeDir(""), ".config/systemd/user", name+".service")
	if helpersUser.IsRoot() {
		serviceFile = filepath.Join("/etc/systemd/system", name+".service")
	}

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromTemplate(name, string(fileContent)), nil
}

func newServiceFromTemplate(name string, template string) Service {
	logging.Debugf("%s: template: %s, from %s", logTag, template, helpersReflect.GetThisFuncName())

	config := Config{
		Name: name,
	}

	for _, line := range strings.Split(template, "\n") {
		if strings.Contains(line, "Description=") {
			lineParts := strings.Split(strings.TrimSpace(line), "Description=")
			if len(lineParts) >= 0 {
				config.Description = lineParts[0]
			}
		} else if strings.Contains(line, "Documentation=") {
			lineParts := strings.Split(strings.TrimSpace(line), "Documentation=")
			if len(lineParts) >= 0 {
				config.Documentation = lineParts[0]
			}
		} else if strings.Contains(line, "ExecStart=") {
			lineParts := strings.Split(strings.TrimSpace(line), "ExecStart=")
			if len(lineParts) >= 0 {
				config.Executable = lineParts[0]
			}
		} else if strings.Contains(line, "WorkingDirectory=") {
			lineParts := strings.Split(strings.TrimSpace(line), "WorkingDirectory=")
			if len(lineParts) >= 0 {
				config.WorkingDirectory = lineParts[0]
			}
		} else if strings.Contains(line, "StandardOutput=") {
			lineParts := strings.Split(strings.TrimSpace(line), "StandardOutput=")
			if len(lineParts) >= 0 {
				config.StdOutPath = lineParts[0]
			}
		} else if strings.Contains(line, "StandardError=") {
			lineParts := strings.Split(strings.TrimSpace(line), "StandardError=")
			if len(lineParts) >= 0 {
				config.StdErrPath = lineParts[0]
			}
		} else if strings.Contains(line, "User=") {
			lineParts := strings.Split(strings.TrimSpace(line), "User=")
			if len(lineParts) >= 0 {
				config.RunAsUser = lineParts[0]
			}
		} else if strings.Contains(line, "Group=") {
			lineParts := strings.Split(strings.TrimSpace(line), "Group=")
			if len(lineParts) >= 0 {
				config.RunAsGroup = lineParts[0]
			}
		}
	}

	return &systemdService{
		config:                 config,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}
}

func (thisRef systemdService) Install() error {
	dir := filepath.Dir(thisRef.filePath())

	// 1.
	logging.Debugf("making sure folder exists: %s, from %s", dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("generating unit file, from %s", helpersReflect.GetThisFuncName())

	fileContent, err := thisRef.fileContentFromConfig()
	if err != nil {
		return err
	}

	if !thisRef.useConfigAsFileContent {
		fileContent = []byte(thisRef.fileContentTemplate)
	}

	logging.Debugf("writing unit to: %s, from %s", thisRef.filePath(), helpersReflect.GetThisFuncName())

	err = ioutil.WriteFile(thisRef.filePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Debugf("wrote unit: %s, from %s", string(fileContent), helpersReflect.GetThisFuncName())

	return nil
}

func (thisRef systemdService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTag, thisRef.config.Name, helpersReflect.GetThisFuncName())

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
	output, err = runSystemCtlCommand("enable", thisRef.config.Name)
	if err != nil {
		if strings.Contains(output, "Failed to enable unit") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Debugf("loading unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("start", thisRef.config.Name)
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
	output, err := runSystemCtlCommand("stop", thisRef.config.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Debugf("disabling unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("disable", thisRef.config.Name)
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
	fileContent, _ := thisRef.fileContentFromDisk()

	result := Info{
		Error:       nil,
		Config:      thisRef.config,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: fileContent,
	}

	output, err := runSystemCtlCommand("status", thisRef.config.Name)
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
		return filepath.Join("/etc/systemd/system", thisRef.config.Name+".service")
	}

	return filepath.Join(helpersUser.HomeDir(""), ".config/systemd/user", thisRef.config.Name+".service")
}

func (thisRef systemdService) fileContentFromConfig() ([]byte, error) {
	if len(thisRef.config.Args) > 0 {
		thisRef.config.Executable = fmt.Sprintf(
			"%s %s",
			thisRef.config.Executable,
			strings.Join(thisRef.config.Args, " "),
		)
	}

	fileTemplate := template.Must(template.New("systemdFile").Parse(`
[Unit]
After=$DependsOn$
Description={{ .Description }}
Documentation={{ .Documentation }}
StartLimitIntervalSec={{ .DelayBeforeRestart }}
StartLimitBurst=0
StartLimitAction=none

[Service]
ExecStart={{ .Executable }}
WorkingDirectory={{ .WorkingDirectory }}
Restart=$Restart$
RestartSec={{ .DelayBeforeRestart }}
Type=simple
{{ if .StdOutPath }}StandardOutput={{ .StdOutPath }}{{ end }}
{{ if .StdErrPath }}StandardError={{ .StdErrPath }}{{ end }}

{{ if .RunAsUser }}User={{ .RunAsUser }}{{ end }}
{{ if .RunAsGroup }}Group={{ .RunAsGroup }}{{ end }}

[Install]
WantedBy=multi-user.target
`))

	var buffer bytes.Buffer
	if err := fileTemplate.Execute(&buffer, thisRef.config); err != nil {
		return nil, err
	}

	fileTemplateAsString := buffer.String()
	fileTemplateAsString = strings.Replace(
		fileTemplateAsString,
		"$DependsOn$",
		strings.Join(thisRef.config.DependsOn, " "),
		1,
	)
	if thisRef.config.Restart {
		fileTemplateAsString = strings.Replace(
			fileTemplateAsString,
			"$Restart$",
			"always",
			1,
		)
	} else {
		fileTemplateAsString = strings.Replace(
			fileTemplateAsString,
			"$Restart$",
			"on-failure",
			1,
		)
	}

	return []byte(fileTemplateAsString), nil
}

func (thisRef systemdService) fileContentFromDisk() ([]byte, error) {
	return ioutil.ReadFile(thisRef.filePath())
}

func runSystemCtlCommand(args ...string) (out string, err error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Debugf("%s: RUN-SYSTEMCTL: systemctl %s, from %s", logTag, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("systemctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-SYSTEMCTL-OUT: output: %s, error: %s, from %s", logTag, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
