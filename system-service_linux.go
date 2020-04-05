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
	helpersFiles "github.com/codemodify/systemkit-helpers-files"
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersUser "github.com/codemodify/systemkit-helpers-os"
	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

var logTag = "SYSTEMD-SERVICE"

// LinuxService - Represents Linux SystemD service
type LinuxService struct {
	command Command
}

// New -
func New(command Command) SystemService {
	logging.Instance().Debugf("%s: config object: %s, from %s", logTag, helpersJSON.AsJSONString(command), helpersReflect.GetThisFuncName())

	return &LinuxService{
		command: command,
	}
}

// Run - is a no-op on Linux based systems
func (thisRef LinuxService) Run() error {
	return nil
}

// Install -
func (thisRef LinuxService) Install(start bool) error {
	dir := filepath.Dir(thisRef.FilePath())

	// 1.
	logging.Instance().Debugf("making sure folder exists: %s, from %s", dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Instance().Debugf("generating unit file, from %s", helpersReflect.GetThisFuncName())

	fileContent, err := thisRef.FileContent()
	if err != nil {
		return err
	}

	logging.Instance().Debugf("writing unit to: %s, from %s", thisRef.FilePath(), helpersReflect.GetThisFuncName())

	err = ioutil.WriteFile(thisRef.FilePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Instance().Debugf("wrote unit: %s, from %s", string(fileContent), helpersReflect.GetThisFuncName())

	// 3.
	if start {
		return thisRef.Start()
	}

	return nil
}

// Start -
func (thisRef LinuxService) Start() error {
	// 1.
	logging.Instance().Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	output, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Instance().Debugf("enabling unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("enable", thisRef.command.Name)
	if err != nil {
		if strings.Contains(output, "Failed to enable unit") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Instance().Debugf("loading unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("start", thisRef.command.Name)
	if err != nil {
		if strings.Contains(output, "Failed to start") && strings.Contains(output, "not found") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

// Restart -
func (thisRef LinuxService) Restart() error {
	if err := thisRef.Stop(); err != nil {
		return err
	}

	return thisRef.Start()
}

// Stop -
func (thisRef LinuxService) Stop() error {
	// 1.
	logging.Instance().Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	_, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Instance().Debugf("stopping unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err := runSystemCtlCommand("stop", thisRef.command.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Instance().Debugf("disabling unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err = runSystemCtlCommand("disable", thisRef.command.Name)
	if err != nil {
		logging.Instance().Warningf("stopping unit file with systemd, from %s", helpersReflect.GetThisFuncName())

		if strings.Contains(output, "Failed to disable") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		} else if strings.Contains(output, "Removed") {
			return nil
		}

		return err
	}

	// 4.
	logging.Instance().Debugf("reloading daemon, from %s", helpersReflect.GetThisFuncName())
	_, err = runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 5.
	logging.Instance().Debugf("running reset-failed, from %s", helpersReflect.GetThisFuncName())
	_, err = runSystemCtlCommand("reset-failed")
	if err != nil {
		return err
	}

	return nil
}

// Uninstall -
func (thisRef LinuxService) Uninstall() error {
	// 1.
	logging.Instance().Debugf("%s: attempting to uninstall: %s, from %s", logTag, thisRef.command.Name, helpersReflect.GetThisFuncName())

	// 2.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 3.
	logging.Instance().Debugf("remove unit file, from %s", helpersReflect.GetThisFuncName())
	err = os.Remove(thisRef.FilePath())
	if e, ok := err.(*os.PathError); ok {
		if os.IsNotExist(e.Err) {
			return nil
		}
	}

	return err
}

// Status -
func (thisRef LinuxService) Status() Status {
	output, err := runSystemCtlCommand("status", thisRef.command.Name)
	if err != nil {
		return Status{
			IsRunning: false,
			PID:       -1,
			Error:     err,
		}
	}

	if strings.Contains(output, "could not be found") {
		return Status{
			IsRunning: false,
			PID:       -1,
			Error:     ErrServiceDoesNotExist,
		}
	}

	status := Status{
		IsRunning: false,
		PID:       -1,
		Error:     nil,
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Main PID") {
			lineParts := strings.Split(strings.TrimSpace(line), " ")
			if len(lineParts) >= 2 {
				status.PID, _ = strconv.Atoi(lineParts[2])
			}
		} else if strings.Contains(line, "Active") {
			if strings.Contains(line, "active (running)") {
				status.IsRunning = true
			}
		}
	}

	return status
}

// Exists -
func (thisRef LinuxService) Exists() bool {
	return helpersFiles.FileOrFolderExists(thisRef.FilePath())
}

// FilePath -
func (thisRef LinuxService) FilePath() string {
	if helpersUser.IsRoot() {
		return filepath.Join("/etc/systemd/system", thisRef.command.Name+".service")
	}

	return filepath.Join(helpersUser.HomeDir(""), ".config/systemd/user", thisRef.command.Name+".service")
}

// FileContent -
func (thisRef LinuxService) FileContent() ([]byte, error) {
	transformedCommand := transformCommandForSaveToDisk(thisRef.command)

	systemDServiceFileTemplate := template.Must(template.New("systemDFile").Parse(`
[Unit]
After=network.target
Description={{ .Description }}
Documentation={{ .DocumentationURL }}

[Service]
ExecStart={{ .Executable }}
WorkingDirectory={{ .WorkingDirectory }}
Restart=on-failure
Type=simple
{{ if .StdOutPath }}StandardOutput={{ .StdOutPath }}{{ end }}
{{ if .StdErrPath }}StandardError={{ .StdErrPath }}{{ end }}

{{ if .RunAsUser }}User={{ .RunAsUser }}{{ end }}
{{ if .RunAsGroup }}Group={{ .RunAsGroup }}{{ end }}

[Install]
WantedBy=multi-user.target
	`))

	var systemDServiceFileTemplateAsBytes bytes.Buffer
	if err := systemDServiceFileTemplate.Execute(&systemDServiceFileTemplateAsBytes, transformedCommand); err != nil {
		return nil, err
	}

	return systemDServiceFileTemplateAsBytes.Bytes(), nil
}

func runSystemCtlCommand(args ...string) (out string, err error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Instance().Debugf("%s: RUN-SYSTEMCTL: systemctl %s, from %s", logTag, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("systemctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Instance().Debugf("%s: RUN-SYSTEMCTL-OUT: output: %s, error: %s, from %s", logTag, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}

func transformCommandForSaveToDisk(command Command) Command {
	if len(command.Args) > 0 {
		command.Executable = fmt.Sprintf("%s %s", command.Executable, strings.Join(command.Args, " "))
	}

	return command
}
