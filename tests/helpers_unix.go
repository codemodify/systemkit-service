// +build !windows

package tests

import (
	"fmt"

	helpersGuid "github.com/codemodify/systemkit-helpers-guid"
	service "github.com/codemodify/systemkit-service"
	"github.com/codemodify/systemkit-service/spec"
)

func CreateService() service.Service {
	return service.NewServiceFromSERVICE(spec.SERVICE{
		Name:             "systemkit-test-service",
		Description:      "SystemKit Test Service",
		Documentation:    "http://systemkit-test-service.com",
		Executable:       "/bin/sleep",
		Args:             []string{"40"},
		WorkingDirectory: "/tmp",
		Environment: map[string]string{
			"TEST-ENV-VAR": "TEST-ENV-VAR-VALUE",
		},
		DependsOn: []spec.ServiceType{
			spec.ServiceNetwork,
		},
		Start: spec.StartConfig{
			AtBoot:         true,
			Restart:        true,
			RestartTimeout: 10,
		},
		Logging: spec.LoggingConfig{
			StdOut: spec.LoggingConfigOut{
				Disabled: true,
			},
			StdErr: spec.LoggingConfigOut{
				Disabled: true,
			},
		},
	})
}

func CreateRandomService() service.Service {
	randomData := helpersGuid.NewGUID()

	serviceCopy := CreateService().Info().Service
	serviceCopy.Name = fmt.Sprintf("systemkit-test-service-%s", randomData)
	serviceCopy.Description = fmt.Sprintf("SystemKit Test Service-%s", randomData)

	return service.NewServiceFromSERVICE(serviceCopy)
}
