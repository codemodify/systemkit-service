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

var logTagSystemV = "SystemV-SERVICE"

type systemvService struct {
	config                 Config
	useConfigAsFileContent bool
	fileContentTemplate    string
}

func newServiceFromConfig_SystemV(config Config) Service {

	config.DependsOn = append(config.DependsOn, "network.target")

	logging.Debugf("%s: config object: %s, from %s", logTagSystemV, helpersJSON.AsJSONString(config), helpersReflect.GetThisFuncName())

	return &systemvService{
		config:                 config,
		useConfigAsFileContent: true,
	}
}

func newServiceFromName_SystemV(name string) (Service, error) {
	serviceFile := filepath.Join("/etc/init.d/", name)

	fileContent, err := ioutil.ReadFile(serviceFile)
	if err != nil {
		return nil, ErrServiceDoesNotExist
	}

	return newServiceFromTemplate_SystemV(name, string(fileContent))
}

func newServiceFromTemplate_SystemV(name string, template string) (Service, error) {
	logging.Debugf("%s: template: %s, from %s", logTagSystemV, template, helpersReflect.GetThisFuncName())

	config := Config{
		Name: name,
		StdOut: LogConfig{
			Disable: true,
		},
		StdErr: LogConfig{
			Disable: true,
		},
	}

	for _, line := range strings.Split(template, "\n") {
		if strings.Contains(line, "# Description:") {
			config.Description = strings.TrimSpace(strings.Replace(line, "# Description:", "", 1))

		} else if strings.Contains(line, "cmd=\"") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "cmd=\"", "", 1))
			parts := strings.Split(cleanLine, " ")
			config.Executable = parts[0]
			config.Args = parts[1:]

		} else if strings.Contains(line, "stdout_log=\"") {
			config.StdOut.Disable = false
			config.StdOut.UseDefault = false
			config.StdOut.Value = strings.TrimSpace(strings.Replace(line, "stdout_log=\"", "", 1))

		} else if strings.Contains(line, "stderr_log=\"") {
			config.StdErr.Disable = false
			config.StdErr.UseDefault = false
			config.StdErr.Value = strings.TrimSpace(strings.Replace(line, "stderr_log=\"", "", 1))

		}
	}

	return &systemvService{
		config:                 config,
		useConfigAsFileContent: false,
		fileContentTemplate:    template,
	}, nil
}

func (thisRef systemvService) Install() error {
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

	err = ioutil.WriteFile(thisRef.filePath(), fileContent, 0755)
	if err != nil {
		return err
	}

	// additional rc.d magic
	for _, i := range [...]string{"2", "3", "4", "5"} {
		if err = os.Symlink(thisRef.filePath(), "/etc/rc"+i+".d/S50"+thisRef.config.Name); err != nil {
			continue
		}
	}
	for _, i := range [...]string{"0", "1", "6"} {
		if err = os.Symlink(thisRef.filePath(), "/etc/rc"+i+".d/K02"+thisRef.config.Name); err != nil {
			continue
		}
	}

	logging.Debugf("wrote unit: %s, from %s", string(fileContent), helpersReflect.GetThisFuncName())

	return nil
}

func (thisRef systemvService) Uninstall() error {
	// 1.
	logging.Debugf("%s: attempting to uninstall: %s, from %s", logTagSystemV, thisRef.config.Name, helpersReflect.GetThisFuncName())

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

func (thisRef systemvService) Start() error {
	// 1.
	logging.Debugf("loading unit file with systemd, from %s", helpersReflect.GetThisFuncName())
	output, err := runServiceCommand(thisRef.config.Name, "start")
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
	logging.Debugf("stopping service, from %s", helpersReflect.GetThisFuncName())
	output, err := runServiceCommand(thisRef.config.Name, "stop")
	if err != nil {
		if strings.Contains(output, "Failed to stop") && strings.Contains(output, "not loaded") {
			return ErrServiceDoesNotExist
		}

		return err
	}

	return nil
}

func (thisRef systemvService) Info() Info {
	fileContent, _ := thisRef.fileContentFromDisk()

	result := Info{
		Error:       nil,
		Config:      thisRef.config,
		IsRunning:   false,
		PID:         -1,
		FilePath:    thisRef.filePath(),
		FileContent: string(fileContent),
	}

	// output, err := runServiceCommand("status", thisRef.config.Name)
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
	return filepath.Join("/etc/init.d/", thisRef.config.Name)
}

func (thisRef systemvService) fileContentFromConfig() ([]byte, error) {
	// for SystemD move everything into config.Executable
	if len(thisRef.config.Args) > 0 {
		thisRef.config.Executable = fmt.Sprintf(
			"%s %s",
			thisRef.config.Executable,
			strings.Join(thisRef.config.Args, " "),
		)
	}

	fileTemplate := template.Must(template.New("systemvFile").Parse(`#!/bin/sh
# For RedHat and cousins:
# chkconfig: - 99 01
# description: {{.Description}}
# processname: {{.Executable}}

### BEGIN INIT INFO
# Provides:          {{.Executable}}
# Required-Start:
# Required-Stop:
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: {{.Name}}
# Description:       {{.Description}}
### END INIT INFO

cmd="{{.Executable}}"

name=$(basename $0)
pid_file="/var/run/$name.pid"
stdout_log="/var/log/$name.log"
stderr_log="/var/log/$name.err"

get_pid() {
    cat "$pid_file"
}

is_running() {
    [ -f "$pid_file" ] && ps $(get_pid) > /dev/null 2>&1
}

case "$1" in
    start)
        if is_running; then
            echo "Already started"
        else
            echo "Starting $name"
            $cmd >> "$stdout_log" 2>> "$stderr_log" &
            echo $! > "$pid_file"
            if ! is_running; then
                echo "Unable to start, see $stdout_log and $stderr_log"
                exit 1
            fi
        fi
    ;;
    stop)
        if is_running; then
            echo -n "Stopping $name.."
            kill $(get_pid)
            for i in {1..10}
            do
                if ! is_running; then
                    break
                fi
                echo -n "."
                sleep 1
            done
            echo
            if is_running; then
                echo "Not stopped; may still be shutting down or shutdown may have failed"
                exit 1
            else
                echo "Stopped"
                if [ -f "$pid_file" ]; then
                    rm "$pid_file"
                fi
            fi
        else
            echo "Not running"
        fi
    ;;
    restart)
        $0 stop
        if is_running; then
            echo "Unable to stop, will not attempt to start"
            exit 1
        fi
        $0 start
    ;;
    status)
        if is_running; then
            echo "Running"
        else
            echo "Stopped"
            exit 1
        fi
    ;;
    *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac
exit 0
`))

	var buffer bytes.Buffer
	if err := fileTemplate.Execute(&buffer, thisRef.config); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (thisRef systemvService) fileContentFromDisk() ([]byte, error) {
	return ioutil.ReadFile(thisRef.filePath())
}

func runServiceCommand(args ...string) (string, error) {
	if !helpersUser.IsRoot() {
		args = append([]string{"--user"}, args...)
	}

	logging.Debugf("%s: RUN-SERVICE: service %s, from %s", logTagSystemV, strings.Join(args, " "), helpersReflect.GetThisFuncName())

	output, err := helpersExec.ExecWithArgs("service", args...)
	errAsString := ""
	if err != nil {
		errAsString = err.Error()
	}

	logging.Debugf("%s: RUN-SERVICE-OUT: output: %s, error: %s, from %s", logTagSystemV, output, errAsString, helpersReflect.GetThisFuncName())

	return output, err
}
