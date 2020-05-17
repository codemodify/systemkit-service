package spec

// LoggingConfigOut -
type LoggingConfigOut struct {
	Disabled   bool   `json:"disabled,omitempty"`   // will disable the service logging
	UseDefault bool   `json:"useDefault,omitempty"` // will set at runtime a default value per platform
	Value      string `json:"value,omitempty"`      // if the two above are false then this will be used as value
}

// LoggingConfig -
type LoggingConfig struct {
	StdOut LoggingConfigOut `json:"stdout,omitempty"`
	StdErr LoggingConfigOut `json:"stderr,omitempty"`
}

// StartConfig -
type StartConfig struct {
	AtBoot         bool `json:"atBoot,omitempty"`         // will indicate if to start at boot
	Restart        bool `json:"restart,omitempty"`        // will disable the restart, similar SystemD: Restart=always/on-failure
	RestartTimeout int  `json:"restartTimeout,omitempty"` // time to wait before restart, similar SystemD: RestartSec=
}

// CredentialsConfig -
type CredentialsConfig struct {
	User  string `json:"user,omitempty"`  //
	Group string `json:"group,omitempty"` //
}

// SERVICE - "common sense" mix of fields from SystemD and LaunchD and rc.d
type SERVICE struct {
	Name                    string                     `json:"name,omitempty"`                    //
	Description             string                     `json:"description,omitempty"`             //
	Documentation           string                     `json:"documentation,omitempty"`           //
	Executable              string                     `json:"executable,omitempty"`              //
	Args                    []string                   `json:"args,omitempty"`                    //
	WorkingDirectory        string                     `json:"workingDirectory,omitempty"`        //
	Environment             map[string]string          `json:"environment,omitempty"`             // similar SystemD: Environment=
	DependsOn               []ServiceType              `json:"dependsOn,omitempty"`               // similar SystemD: Requires=
	DependsOnOverrideByOS   map[OsType][]ServiceType   `json:"dependsOnOverrideByOS,omitempty"`   //
	DependsOnOverrideByInit map[InitType][]ServiceType `json:"dependsOnOverrideByInit,omitempty"` //
	Start                   StartConfig                `json:"start,omitempty"`                   //
	Logging                 LoggingConfig              `json:"logging,omitempty"`                 //
	Credentials             CredentialsConfig          `json:"credentials,omitempty"`             //

	OnStopDelegate func() `json:"-"` //
}
