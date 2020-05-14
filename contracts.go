package service

import (
	"errors"
)

// ErrServiceDoesNotExist -
var ErrServiceDoesNotExist = errors.New("Service does not exist")

// ErrServiceConfigError -
var ErrServiceConfigError = errors.New("Service config error")

// LogConfig -
type LogConfig struct {
	Disable    bool   `json:"disable"`    // will disable the service logging
	UseDefault bool   `json:"useDefault"` // will set at runtime a default value per platform
	Value      string `json:"value"`      // if the two above are false then this will be used as value
}

// Config - "common sense" mix of fields from SystemD and LaunchD.
// Some fields are ignored on platforms where it does not make sense.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html
// https://www.freedesktop.org/software/systemd/man/systemd.unit.html
// https://www.manpagez.com/man/5/launchd.plist
type Config struct {
	Name               string    `json:"name"`               //
	Description        string    `json:"description"`        //
	Documentation      string    `json:"documentation"`      //
	Executable         string    `json:"executable"`         //
	Args               []string  `json:"args"`               //
	WorkingDirectory   string    `json:"workingDirectory"`   //
	Environment        string    `json:"environment"`        // similar SystemD: Environment=
	DependsOn          []string  `json:"dependsOn"`          // similar SystemD: Requires=
	Restart            bool      `json:"restart"`            // similar SystemD: Restart=always/on-failure
	DelayBeforeRestart int       `json:"delayBeforeRestart"` // similar SystemD: RestartSec=
	StdOut             LogConfig `json:"stdOut"`             //
	StdErr             LogConfig `json:"stdErr"`             //
	RunAsUser          string    `json:"runAsUser"`          //
	RunAsGroup         string    `json:"runAsGroup"`         //
	OnStopDelegate     func()    `json:"-"`                  //
}

// Info -
type Info struct {
	Error       error  `json:"error"` // can be ErrServiceDoesNotExist
	Config      Config `json:"config"`
	IsRunning   bool   `json:"isRunning"`
	PID         int    `json:"pid"`
	FilePath    string `json:"filePath"`
	FileContent string `json:"fileContent"`
}
