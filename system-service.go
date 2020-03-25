package service

// SystemService - represents a generic system service configuration
type SystemService interface {
	Run() error
	Install(start bool) error
	Start() error
	Restart() error
	Stop() error
	Uninstall() error
	Status() Status
	Exists() bool
	FilePath() string
	FileContent() ([]byte, error)
}

// Command - What to execute as service
// These fields are a "common sense" mix of fields from SystemD and LaunchD.
// Some may be ignored on one or other platform but the implemetnation will
// try the max possible to respect the requested
type Command struct {
	Name                string // usually this will be the file name
	DisplayLabel        string
	Description         string
	DocumentationURL    string
	Executable          string
	Args                []string
	WorkingDirectory    string
	Debug               bool
	KeepAlive           bool
	RunAtLoad           bool
	StdOutPath          string
	StdErrPath          string
	StartDelayInSeconds int
	RunAsUser           string
	RunAsGroup          string
	OnStopDelegate      func() `json:"-"`
}

// Status - is a generic representation of the service running on the system
type Status struct {
	IsRunning bool
	PID       int
	Error     error
}
