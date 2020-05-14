package service

import (
	"errors"
)

// ErrServiceDoesNotExist -
var ErrServiceDoesNotExist = errors.New("Service does not exist")

// ErrServiceConfigError -
var ErrServiceConfigError = errors.New("Service config error")

// ErrServiceUnsupportedRequest -
var ErrServiceUnsupportedRequest = errors.New("Service unsupported request")

// LogConfig -
type LogConfig struct {
	Disable    bool   `json:"disable,omitempty"`    // will disable the service logging
	UseDefault bool   `json:"useDefault,omitempty"` // will set at runtime a default value per platform
	Value      string `json:"value,omitempty"`      // if the two above are false then this will be used as value
}

// Config - "common sense" mix of fields from SystemD and LaunchD.
// Some fields are ignored on platforms where it does not make sense.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html
// https://www.freedesktop.org/software/systemd/man/systemd.unit.html
// https://www.manpagez.com/man/5/launchd.plist
type Config struct {
	Name               string    `json:"name,omitempty"`               //
	Description        string    `json:"description,omitempty"`        //
	Documentation      string    `json:"documentation,omitempty"`      //
	Executable         string    `json:"executable,omitempty"`         //
	Args               []string  `json:"args,omitempty"`               //
	WorkingDirectory   string    `json:"workingDirectory,omitempty"`   //
	Environment        string    `json:"environment,omitempty"`        // similar SystemD: Environment=
	DependsOn          []string  `json:"dependsOn,omitempty"`          // similar SystemD: Requires=
	Restart            bool      `json:"restart,omitempty"`            // similar SystemD: Restart=always/on-failure
	DelayBeforeRestart int       `json:"delayBeforeRestart,omitempty"` // similar SystemD: RestartSec=
	StdOut             LogConfig `json:"stdOut,omitempty"`             //
	StdErr             LogConfig `json:"stdErr,omitempty"`             //
	RunAsUser          string    `json:"runAsUser,omitempty"`          //
	RunAsGroup         string    `json:"runAsGroup,omitempty"`         //
	OnStopDelegate     func()    `json:"-"`                            //
}

// Info -
type Info struct {
	Error       error  `json:"-"`
	Config      Config `json:"config,omitempty"`
	IsRunning   bool   `json:"isRunning"`
	PID         int    `json:"pid,omitempty"`
	FilePath    string `json:"filePath,omitempty"`
	FileContent string `json:"fileContent,omitempty"`
}
