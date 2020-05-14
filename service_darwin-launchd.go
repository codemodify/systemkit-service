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
	"github.com/groob/plist"
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

	if config.StdOut.UseDefault {
		config.StdOut.Value = filepath.Join(logDir, config.Name+".stdout.log")
	}

	if config.StdErr.UseDefault {
		config.StdErr.Value = filepath.Join(logDir, config.Name+".stderr.log")
	}

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

	return newServiceFromTemplate(name, string(fileContent))
}

func newServiceFromTemplate(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTag, template, helpersReflect.GetThisFuncName())

	config := Config{
		Name: name,
		StdOut: LogConfig{
			Disable: true,
		},
		StdErr: LogConfig{
			Disable: true,
		},
	}

	var plistData struct {
		Label             *string  `plist:"Label"`
		ProgramArguments  []string `plist:"ProgramArguments"`
		StandardOutPath   *string  `plist:"StandardOutPath"`
		StandardErrorPath *string  `plist:"StandardErrorPath"`
		KeepAlive         *bool    `plist:"KeepAlive"`
		WorkingDirectory  *string  `plist:"WorkingDirectory"`
	}

	if err := plist.NewXMLDecoder(strings.NewReader(template)).Decode(&plistData); err != nil {
		logging.Errorf("%s: error parsing PLIST: %s, from %s", logTag, err.Error(), helpersReflect.GetThisFuncName())
	} else {
		if len(plistData.ProgramArguments) > 0 {
			config.Executable = plistData.ProgramArguments[0]
			config.Args = plistData.ProgramArguments[1:]
		}
		if plistData.StandardOutPath != nil {
			config.StdOut.Disable = false
			config.StdOut.UseDefault = false
			config.StdOut.Value = *plistData.StandardOutPath
		}
		if plistData.StandardErrorPath != nil {
			config.StdErr.Disable = false
			config.StdErr.UseDefault = false
			config.StdErr.Value = *plistData.StandardErrorPath
		}
		config.Restart = *plistData.KeepAlive
		config.WorkingDirectory = *plistData.WorkingDirectory
	}

	return &launchdService{
		config:                 config,
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
	fileContent, err := thisRef.fileContentFromConfig()
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
	runLaunchCtlCommand("remove", thisRef.config.Name)
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

	runLaunchCtlCommand("start", thisRef.config.Name)
	return nil
}

func (thisRef launchdService) Stop() error {
	runLaunchCtlCommand("stop", thisRef.config.Name)
	output, err := runLaunchCtlCommand("unload", thisRef.filePath())
	if strings.Contains(output, "Could not find specified service") {
		return ErrServiceDoesNotExist
	}

	return err
}

func (thisRef launchdService) Info() Info {
	fileContent, fileContentErr := thisRef.fileContentFromDisk()

	result := Info{
		Error:       nil,
		Config:      thisRef.config,
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

		if chunks[2] == thisRef.config.Name {
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
		return filepath.Join("/Library/LaunchDaemons", thisRef.config.Name+".plist")
	}

	return filepath.Join(helpersUser.HomeDir(""), "Library/LaunchAgents", thisRef.config.Name+".plist")
}

func (thisRef launchdService) fileContentFromConfig() ([]byte, error) {
	// for LaunchD move everything into config.Args
	args := []string{thisRef.config.Executable}

	if len(thisRef.config.Args) > 0 {
		args = append(args, thisRef.config.Args...)
	}

	thisRef.config.Args = args

	// run the template
	fileTemplate := template.Must(template.New("launchdFile").Parse(`
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

		{{ if eq .StdOut.Disable false}}
		<key>StandardOutPath</key>
		<string>{{ .StdOut.Value }}</string>
		{{ end }}

		{{ if eq .StdErr.Disable false}}
		<key>StandardErrorPath</key>
		<string>{{ .StdErr.Value }}</string>
		{{ end }}

		<key>KeepAlive</key>
		<{{ .Restart }}/>
		<key>RunAtLoad</key>
		<true/>

		<key>WorkingDirectory</key>
		<string>{{ .WorkingDirectory }}</string>
	</dict>
</plist>
`))

	var buffer bytes.Buffer
	if err := fileTemplate.Execute(&buffer, thisRef.config); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (thisRef launchdService) fileContentFromDisk() ([]byte, error) {
	return ioutil.ReadFile(thisRef.filePath())
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
