package service

import (
	spec "github.com/codemodify/systemkit-service-spec"
)

// Installer - installs and removes a service
type Installer interface {
	Install() error
	Uninstall() error
}

// Controller - starts and stops a service
type Controller interface {
	Start() error
	Stop() error
}

// Describer - gets info about a service
type Describer interface {
	Info() Info
}

// Service -
type Service interface {
	Installer
	Controller
	Describer
}

// NewServiceFromSERVICE -
func NewServiceFromSERVICE(serviceSpec spec.SERVICE) Service {
	return newServiceFromSERVICE(serviceSpec)
}

// NewServiceFromName -
func NewServiceFromName(name string) (Service, error) {
	return newServiceFromName(name)
}

// NewServiceFromPlatformTemplate -
func NewServiceFromPlatformTemplate(name string, template string) (Service, error) {
	return newServiceFromPlatformTemplate(name, template)
}

// Info -
type Info struct {
	Error       error        `json:"-"`
	Service     spec.SERVICE `json:"config,omitempty"`
	IsRunning   bool         `json:"isRunning"`
	PID         int          `json:"pid,omitempty"`
	FilePath    string       `json:"filePath,omitempty"`
	FileContent string       `json:"fileContent,omitempty"`
}
