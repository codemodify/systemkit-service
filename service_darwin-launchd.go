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
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersUser "github.com/codemodify/systemkit-helpers-os"
	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

var logTag = "LAUNCHD-SERVICE"

type launchdService struct {
	config                 Config
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromConfig(config Config) Service {
	// override some values - platform specific
	// https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
	logDir := filepath.Join(helpersUser.HomeDir(""), "Library/Logs", config.Name)
	if helpersUser.IsRoot() {
		logDir = filepath.Join("/Library/Logs", config.Name)
	}

	config.Args = append([]string{config.Executable}, config.Args...)
	config.StdOutPath = filepath.Join(logDir, config.Name+".stdout.log")
	config.StdErrPath = filepath.Join(logDir, config.Name+".stderr.log")

	logging.Debugf("%s: config object: %s, from %s", logTag, helpersJSON.AsJSONString(config), helpersReflect.GetThisFuncName())

	launchdService := &launchdService{
		config:                 config,
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

	return &launchdService{
		config:                 config,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}
}

func (thisRef launchdService) Install() error {
	dir := filepath.Dir(thisRef.FilePath())

	// 1.
	logging.Debugf("%s: making sure folder exists: %s, from %s", logTag, dir, helpersReflect.GetThisFuncName())
	os.MkdirAll(dir, os.ModePerm)

	// 2.
	logging.Debugf("%s: generating plist file, from %s", logTag, helpersReflect.GetThisFuncName())
	fileContent, err := thisRef.fileContentFromConfig()
	if err != nil {
		return err
	}

	logging.Debugf("%s: writing plist to: %s, from %s", logTag, thisRef.FilePath(), helpersReflect.GetThisFuncName())
	err = ioutil.WriteFile(thisRef.FilePath(), fileContent, 0644)
	if err != nil {
		return err
	}

	logging.Debugf("%s: wrote unit: %s, from %s", logTag, string(fileContent), helpersReflect.GetThisFuncName())

	// 3.
	if start {
		return thisRef.Start()
	}

	return nil
}

func (thisRef launchdService) Uninstall() error {
	// 1.
	err := thisRef.Stop()
	if err != nil && !helpersErrors.Is(err, ErrServiceDoesNotExist) {
		return err
	}

	// 2.
	logging.Debugf("%s: remove plist file: %s, from %s", logTag, thisRef.FilePath(), helpersReflect.GetThisFuncName())
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
	runLaunchCtlCommand("remove", thisRef.config.Name)
	return nil
}

func (thisRef launchdService) Start() error {
	// 1.
	output, _ := runLaunchCtlCommand("load", "-w", thisRef.FilePath())
	if strings.Contains(output, "No such file or directory") {
		return ErrServiceDoesNotExist
	}

	if strings.Contains(output, "service already loaded") {
		logging.Debugf("service already loaded, from %s", helpersReflect.GetThisFuncName())

		return nil
	}

	runLaunchCtlCommand("start", thisRef.config.Name)
	return nil
}

func (thisRef launchdService) Stop() error {
	runLaunchCtlCommand("stop", thisRef.config.Name)
	output, err := runLaunchCtlCommand("unload", thisRef.FilePath())
	if strings.Contains(output, "Could not find specified service") {
		return ErrServiceDoesNotExist
	}

	return err
}

func (thisRef launchdService) Info() Info {
	output, err := runLaunchCtlCommand("list")
	if err != nil {
		logging.Errorf("error getting launchctl status: %s, from %s", err, helpersReflect.GetThisFuncName())
		return Status{
			IsRunning: false,
			PID:       -1,
			Error:     err,
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if thisRef.config.Name == "" {
		return Status{}
	}

	status := Status{
		IsRunning: false,
		PID:       -1,
		Error:     nil,
	}
	for _, line := range lines {
		chunks := strings.Split(line, "\t")

		if chunks[2] == thisRef.config.Name {
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

func (thisRef launchdService) filePath() string {
	if helpersUser.IsRoot() {
		return filepath.Join("/Library/LaunchDaemons", thisRef.config.Name+".plist")
	}

	return filepath.Join(helpersUser.HomeDir(""), "Library/LaunchAgents", thisRef.config.Name+".plist")
}

func (thisRef launchdService) fileContentFromConfig() ([]byte, error) {
	plistTemplate := template.Must(template.New("launchdFile").Parse(`
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
	if err := plistTemplate.Execute(&plistTemplateBytes, thisRef.config); err != nil {
		return nil, err
	}

	return plistTemplateBytes.Bytes(), nil
}

func runLaunchCtlCommand(args ...string) (out string, err error) {
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
