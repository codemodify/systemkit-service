// +build !windows
// +build !darwin

package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	helpersExec "github.com/codemodify/systemkit-helpers-os"
	helpersUser "github.com/codemodify/systemkit-helpers-os"
	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

var logTagUpstart = "UpStart-SERVICE"

type upstartService struct {
	config                 Config
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromConfig_Upstart(config Config) Service {

	config.DependsOn = append(config.DependsOn, "network.target")

	logging.Debugf("%s: config object: %s, from %s", logTagUpstart, helpersJSON.AsJSONString(config), helpersReflect.GetThisFuncName())

	return &upstartService{
		config:                 config,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName_Upstart(name string) (Service, error) {
	serviceFile := filepath.Join("/etc/init/", name+".conf")

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromTemplate_Upstart(name, string(fileContent))
}

func newServiceFromTemplate_Upstart(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTagUpstart, template, helpersReflect.GetThisFuncName())

	config := Config{
		Name: name,
		StdOut: LogConfig{
			Disable: true,
		},
		StdErr: LogConfig{
			Disable: true,
		},
	}

	for lineIndex, line := range strings.Split(template, "\n") {
		if strings.Contains(line, "# ") && lineIndex == 0 {
			config.Description = strings.TrimSpace(strings.Replace(line, "# ", "", 1))

		} else if strings.Contains(line, "exec ") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "exec ", "", 1))
			parts := strings.Split(cleanLine, " ")
			config.Executable = parts[0]
			config.Args = parts[1:]

		}
	}

	return &upstartService{
		config:                 config,
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

func (thisRef upstartService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTagUpstart, thisRef.config.Name, helpersReflect.GetThisFuncName())

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
	output, err := runInitctlCommand("start", thisRef.config.Name)
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
	output, err := runInitctlCommand("stop", thisRef.config.Name)
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef upstartService) Info() Info {
	fileContent, _ := thisRef.fileContentFromDisk()

	result := Info{
		Error:       nil,
		Config:      thisRef.config,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	// output, err := runInitctlCommand("status", thisRef.config.Name)
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
	return filepath.Join("/etc/init/", thisRef.config.Name+".conf")
}

func (thisRef upstartService) fileContentFromConfig() ([]byte, error) {
	// for SystemD move everything into config.Executable
	if len(thisRef.config.Args) > 0 {
		thisRef.config.Executable = fmt.Sprintf(
			"%s %s",
			thisRef.config.Executable,
			strings.Join(thisRef.config.Args, " "),
		)
	}

	fileTemplate := template.Must(template.New("upstartFile").Parse(`# {{.Description}}

description     "{{.Name}}"

start on filesystem or runlevel [2345]
stop on runlevel [!2345]

#setuid username

# stop the respawn is process fails to start 5 times within 5 minutes
respawn
respawn limit 5 300
umask 022

console none

pre-start script
    test -x {{.Executable}} || { stop; exit 0; }
end script

# Start
exec {{.Executable}}
`))

	var buffer bytes.Buffer
	if err := fileTemplate.Execute(&buffer, thisRef.config); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (thisRef upstartService) fileContentFromDisk() ([]byte, error) {
	return ioutil.ReadFile(thisRef.filePath())
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
