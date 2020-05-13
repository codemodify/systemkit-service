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
		Executable:         "htop",
		Args:               []string{""},
		WorkingDirectory:   "/tmp",
		Restart:            true,
		DelayBeforeRestart: 10,
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
		StdOutPath:       "null",
		RunAsUser:        "user",
	})
}

func createRemoteitService() service.Service {
	return service.NewServiceFromConfig(service.Config{
		Name:             "it.remote.cli",
		Description:      "it.remote.cli",
		Executable:       "/Users/nicolae/Downloads/remoteit_mac-osx_x86_64",
		Args:             []string{"watch", "-v", "-c", "/etc/remoteit/config.json"},
		WorkingDirectory: "",
		StdOutPath:       "null",
		RunAsUser:        "user",
	})
}
