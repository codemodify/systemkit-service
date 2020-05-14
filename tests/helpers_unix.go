// +build !windows

package tests

import (
	"fmt"

	helpersGuid "github.com/codemodify/systemkit-helpers-guid"
	service "github.com/codemodify/systemkit-service"
)

func createService() service.Service {
	return service.NewServiceFromConfig(service.Config{
		Name:               "systemkit-test-service",
		Description:        "SystemKit Test Service",
		Executable:         "/bin/sleep",
		Args:               []string{"40"},
		WorkingDirectory:   "/tmp",
		Restart:            true,
		DelayBeforeRestart: 10,
		StdOut: service.LogConfig{
			Disable: true,
		},
		StdErr: service.LogConfig{
			Disable: true,
		},
	})
}

func createRandomService() service.Service {
	randomData := helpersGuid.NewGUID()

	return service.NewServiceFromConfig(service.Config{
		Name:             fmt.Sprintf("systemkit-test-service-%s", randomData),
		Description:      fmt.Sprintf("SystemKit Test Service-%s", randomData),
		Executable:       "htop",
		Args:             []string{""},
		WorkingDirectory: "/tmp",
		StdOut: service.LogConfig{
			Disable: true,
		},
		StdErr: service.LogConfig{
			Disable: true,
		},
	})
}

func CreateRemoteitService() service.Service {
	return service.NewServiceFromConfig(service.Config{
		Name:             "it.remote.cli",
		Description:      "it.remote.cli",
		Executable:       "/Users/nicolae/Downloads/remoteit_mac-osx_x86_64",
		Args:             []string{"watch", "-v", "-c", "/etc/remoteit/config.json"},
		WorkingDirectory: "",
		StdOut: service.LogConfig{
			Disable: true,
		},
		StdErr: service.LogConfig{
			Disable: true,
		},
	})
}
