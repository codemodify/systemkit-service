// +build darwin

package service

import (
	"bytes"
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

var logTag = "MACOS-SYSTEM-SERVICE"

// MacOSService - Represents Mac OS Service service
type MacOSService struct {
	command Command
}

// New -
func New(command Command) SystemService {
	// override some values - platform specific
	// https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
	logDir := filepath.Join(helpersUser.HomeDir(""), "Library/Logs", command.Name)
	if helpersUser.IsRoot() {
		logDir = filepath.Join("/Library/Logs", command.Name)
	}

	command.Args = append([]string{command.Executable}, command.Args...)
	command.KeepAlive = true
	command.RunAtLoad = true
	command.StdOutPath = filepath.Join(logDir, command.Name+".stdout.log")
	command.StdErrPath = filepath.Join(logDir, command.Name+".stderr.log")

	macOSService := &MacOSService{
		command: command,
	}

	logging.Instance().Debugf("%s: config object: %s, from %s", logTag, helpersJSON.AsJSONString(command), helpersReflect.GetThisFuncName())

	return macOSService
}

// Run - is a no-op on Mac based systems
func (thisRef MacOSService) Run() error {
	return nil
}

// Install -
func (thisRef MacOSService) Install(start bool) error {
	dir := filepath.Dir(thisRef.FilePath())

	// 1.
	logging.Instance().Debugf("%s: making sure folder exists: %s, from %s", logTag, dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Instance().Debugf("%s: generating plist file, from %s", logTag, helpersReflect.GetThisFuncName())
	fileContent, err := thisRef.FileContent()
	if err != nil {
		return err
	}

	logging.Instance().Debugf("%s: writing plist to: %s, from %s", logTag, thisRef.FilePath(), helpersReflect.GetThisFuncName())
	err = ioutil.WriteFile(thisRef.FilePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Instance().Debugf("%s: wrote unit: %s, from %s", logTag, string(fileContent), helpersReflect.GetThisFuncName())

	// 3.
	if start {
		return thisRef.Start()
	}

	return nil
}

// Start -
func (thisRef MacOSService) Start() error {
	// 1.
	output, _ := runLaunchCtlCommand("load", "-w", thisRef.FilePath())
	if strings.Contains(output, "No such file or directory") {
		return ErrServiceDoesNotExist
	}

	if strings.Contains(output, "service already loaded") {
		logging.Instance().Debugf("service already loaded, from %s", helpersReflect.GetThisFuncName())

		return nil
	}

	runLaunchCtlCommand("start", thisRef.command.Name)
	return nil
}

// Restart -
func (thisRef MacOSService) Restart() error {
	if err := thisRef.Stop(); err != nil {
		return err
	}

	return thisRef.Start()
}

// Stop -
func (thisRef MacOSService) Stop() error {
	runLaunchCtlCommand("stop", thisRef.command.Name)
	output, err := runLaunchCtlCommand("unload", thisRef.FilePath())
	if strings.Contains(output, "Could not find specified service") {
		return ErrServiceDoesNotExist
	}

	return err
}

// Uninstall -
func (thisRef MacOSService) Uninstall() error {
	// 1.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 2.
	logging.Instance().Debugf("%s: remove plist file: %s, from %s", logTag, thisRef.FilePath(), helpersReflect.GetThisFuncName())
	err = os.Remove(thisRef.FilePath())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such file or directory") {
			return nil
		}

		return err
	}

	// INFO: ignore the return value as is it is barely defined by the docs
	// what the expected behavior would be. The previous stop and remove the "plist" file
	// will uninstall the service.
	runLaunchCtlCommand("remove", thisRef.command.Name)
	return nil
}

// Status -
func (thisRef MacOSService) Status() Status {
	output, err := runLaunchCtlCommand("list")
	if err != nil {
		logging.Instance().Errorf("error getting launchctl status: %s, from %s", err, helpersReflect.GetThisFuncName())
		return Status{
			IsRunning: false,
			PID:       -1,
			Error:     err,
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if thisRef.command.Name == "" {
		return Status{}
	}

	status := Status{
		IsRunning: false,
		PID:       -1,
		Error:     nil,
	}
	for _, line := range lines {
		chunks := strings.Split(line, "\t")

		if chunks[2] == thisRef.command.Name {
			if chunks[0] != "-" {
				pid, err := strconv.Atoi(chunks[0])
				if err != nil {
					return status
				}
				status.PID = pid
			}

			if status.PID != -1 {
				status.IsRunning = true
			}

			break
		}
	}

	return status
}

// Exists -
func (thisRef MacOSService) Exists() bool {
	return helpersFiles.FileOrFolderExists(thisRef.FilePath())
}

// FilePath -
func (thisRef MacOSService) FilePath() string {
	if helpersUser.IsRoot() {
		return filepath.Join("/Library/LaunchDaemons", thisRef.command.Name+".plist")
	}

	return filepath.Join(helpersUser.HomeDir(""), "Library/LaunchAgents", thisRef.command.Name+".plist")
}

// FileContent -
func (thisRef MacOSService) FileContent() ([]byte, error) {
	plistTemplate := template.Must(template.New("launchdConfig").Parse(`
<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
<plist version='1.0'>
	<dict>
		<key>Label</key>
		<string>{{ .Name }}</string>

		<key>ProgramArguments</key>
		<array>{{ range $arg := .Args }}
			<string>{{ $arg }}</string>{{ end }}
		</array>

		<key>StandardOutPath</key>
		<string>{{ .StdOutPath }}</string>

		<key>StandardErrorPath</key>
		<string>{{ .StdErrPath }}</string>

		<key>KeepAlive</key>
		<{{ .KeepAlive }}/>
		<key>RunAtLoad</key>
		<{{ .RunAtLoad }}/>

		<key>WorkingDirectory</key>
		<string>{{ .WorkingDirectory }}</string>
	</dict>
</plist>
	`))

	var plistTemplateBytes bytes.Buffer
	if err := plistTemplate.Execute(&plistTemplateBytes, thisRef.command); err != nil {
		return nil, err
	}

	return plistTemplateBytes.Bytes(), nil
}

func runLaunchCtlCommand(args ...string) (out string, err error) {
	// if !helpersUser.IsRoot() {
	// 	args = append([]string{"--user"}, args...)
	// }

	logging.Instance().Debugf("%s: RUN-LAUNCHCTL: launchctl %s, from %s", logTag, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("launchctl", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Instance().Debugf("%s: RUN-LAUNCHCTL-OUT: output: %s, error: %s, from %s", logTag, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
