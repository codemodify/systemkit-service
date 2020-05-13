package service

import (
	"errors"
)

// ErrServiceDoesNotExist -
var ErrServiceDoesNotExist = errors.New("Service does not exist")

// Config - "common sense" mix of fields from SystemD and LaunchD.
// Some fields are ignored on platforms where it does not make sense.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html#
// https://www.freedesktop.org/software/systemd/man/systemd.unit.html#
type Config struct {
	Name               string   //
	Description        string   //
	Documentation      string   //
	Executable         string   //
	Args               []string //
	WorkingDirectory   string   //
	Environment        string   // similar SystemD: Environment=
	DependsOn          []string // similar SystemD: Requires=
	Restart            bool     // similar SystemD: Restart=always/on-failure
	DelayBeforeRestart int      // similar SystemD: RestartSec=
	StdOutPath         string   //
	StdErrPath         string   //
	RunAsUser          string   //
	RunAsGroup         string   //
	OnStopDelegate     func()   `json:"-"`
}

// Info -
type Info struct {
	Error       error // can be ErrServiceDoesNotExist
	Config      Config
	IsRunning   bool
	PID         int
	FilePath    string
	FileContent []byte
}
