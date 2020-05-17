// +build windows

package tests

import (
	"fmt"

	helpersGuid "github.com/codemodify/systemkit-helpers-guid"
	service "github.com/codemodify/systemkit-service"
)

func CreateService() service.Service {
	return service.NewServiceFromConfig(service.Config{
		Name:        "systemkit-test-service",
		Description: "SystemKit Test Service",
		// Executable:       "C:\\Program Files (x86)\\Plex\\Plex Media Server\\Plex Update Service.exe",
		Executable:       "C:\\Windows\\notepad.exe",
		Args:             []string{"aaaaaaaaaaa"},
		WorkingDirectory: "C:\\Windows",
		StdOut: service.LogConfig{
			Disable: true,
		},
		StdErr: service.LogConfig{
			Disable: true,
		},
	})
}

func CreateRandomService() service.Service {
	randomData := helpersGuid.NewGUID()

	s := CreateService()

	config := s.Info().Config
	config.Name = fmt.Sprintf("%s-%s", config.Name, randomData)
	config.Description = fmt.Sprintf("%s-%s", config.Description, randomData)

	return service.NewServiceFromConfig(config)
}
