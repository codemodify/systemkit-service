package encoders

import "github.com/codemodify/systemkit-service/spec"

func newEmptySERVICE() spec.SERVICE {
	return spec.SERVICE{
		Name:                    "",
		Description:             "",
		Documentation:           "",
		Executable:              "",
		Args:                    []string{},
		WorkingDirectory:        "",
		Environment:             map[string]string{},
		DependsOn:               []spec.ServiceType{},
		DependsOnOverrideByOS:   map[spec.OsType][]spec.ServiceType{},
		DependsOnOverrideByInit: map[spec.InitType][]spec.ServiceType{},
		Start:                   spec.StartConfig{},
		Logging: spec.LoggingConfig{
			StdOut: spec.LoggingConfigOut{
				Disabled: true,
			},
			StdErr: spec.LoggingConfigOut{
				Disabled: true,
			},
		},
		Credentials: spec.CredentialsConfig{},
	}
}
