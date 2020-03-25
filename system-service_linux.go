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

	helpersErrors "github.com/codemodify/systemkit-helpers"
	helpersExec "github.com/codemodify/systemkit-helpers"
	helpersFiles "github.com/codemodify/systemkit-helpers"
	helpersJSON "github.com/codemodify/systemkit-helpers"
	helpersReflect "github.com/codemodify/systemkit-helpers"
	helpersUser "github.com/codemodify/systemkit-helpers"
	logging "github.com/codemodify/systemkit-logging"
	loggingC "github.com/codemodify/systemkit-logging/contracts"
)

var logTag = "SYSTEMD-SERVICE"

// LinuxService - Represents Linux SystemD service
type LinuxService struct {
	command Command
}

// New -
func New(command Command) SystemService {
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: config object: %s ", logTag, helpersJSON.AsJSONString(command)),
	})

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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprint("making sure folder exists: ", dir),
	})
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprint("generating unit file"),
	})
	fileContent, err := thisRef.FileContent()
	if err != nil {
		return err
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprint("writing unit to: ", thisRef.FilePath()),
	})
	err = ioutil.WriteFile(thisRef.FilePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("wrote unit: %s", string(fileContent)),
	})

	// 3.
	if start {
		return thisRef.Start()
	}

	return nil
}

// Start -
func (thisRef LinuxService) Start() error {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "reloading daemon",
	})
	output, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "enabling unit file with systemd",
	})
	output, err = runSystemCtlCommand("enable", thisRef.command.Name)
	if err != nil {
		if strings.Contains(output, "Failed to enable unit") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "loading unit file with systemd",
	})
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
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "reloading daemon",
	})
	_, err := runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 2.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "stopping unit file with systemd",
	})
	output, err := runSystemCtlCommand("stop", thisRef.command.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	// 3.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "disabling unit file with systemd",
	})
	output, err = runSystemCtlCommand("disable", thisRef.command.Name)
	if err != nil {
		logging.Instance().LogWarningWithFields(loggingC.Fields{
			"method":  helpersReflect.GetThisFuncName(),
			"message": "stopping unit file with systemd",
		})

		if strings.Contains(output, "Failed to disable") && strings.Contains(output, "does not exist") {
			return ErrServiceDoesNotExist
		} else if strings.Contains(output, "Removed") {
			return nil
		}

		return err
	}

	// 4.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "reloading daemon",
	})
	_, err = runSystemCtlCommand("daemon-reload")
	if err != nil {
		return err
	}

	// 5.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "running reset-failed",
	})
	_, err = runSystemCtlCommand("reset-failed")
	if err != nil {
		return err
	}

	return nil
}

// Uninstall -
func (thisRef LinuxService) Uninstall() error {
	// 1.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: attempting to uninstall: %s", logTag, thisRef.command.Name),
	})

	// 2.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 3.
	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": "remove unit file",
	})
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

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: RUN-SYSTEMCTL: systemctl %s", logTag, strings.Join(args, " ")),
	})

	output, err := helpersExec.ExecWithArgs("systemctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Instance().LogDebugWithFields(loggingC.Fields{
		"method":  helpersReflect.GetThisFuncName(),
		"message": fmt.Sprintf("%s: RUN-SYSTEMCTL-OUT: output: %s, error: %s", logTag, output, errAsString),
	})

	return output, err
}

func transformCommandForSaveToDisk(command Command) Command {
	if len(command.Args) > 0 {
		command.Executable = fmt.Sprintf("%s %s", command.Executable, strings.Join(command.Args, " "))
	}

	return command
}
