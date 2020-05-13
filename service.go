package service

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

// NewServiceFromConfig -
func NewServiceFromConfig(config Config) Service {
	return newServiceFromConfig(config)
}

// NewServiceFromName -
func NewServiceFromName(name string) (Service, error) {
	return newServiceFromName(name)
}

// NewServiceFromTemplate -
func NewServiceFromTemplate(name string, template string) Service {
	return newServiceFromTemplate(name, template)
}
