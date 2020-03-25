// +build !windows

package tests

import (
	"fmt"

	helpersGuid "github.com/codemodify/systemkit-helpers"
	service "github.com/codemodify/systemkit-service"
)

func createService() service.SystemService {
	return service.New(service.Command{
		Name:             "systemkit-test-service",
		DisplayLabel:     "SystemKit Test Service",
		Description:      "SystemKit Test Service",
		DocumentationURL: "",
		Executable:       "htop",
		Args:             []string{""},
		WorkingDirectory: "/tmp",
		StdOutPath:       "null",
		RunAsUser:        "user",
	})
}

func createRandomService() service.SystemService {
	randomData := helpersGuid.NewGUID()

	return service.New(service.Command{
		Name:             fmt.Sprintf("systemkit-test-service-%s", randomData),
		DisplayLabel:     fmt.Sprintf("SystemKit Test Service-%s", randomData),
		Description:      fmt.Sprintf("SystemKit Test Service-%s", randomData),
		DocumentationURL: "",
		Executable:       "htop",
		Args:             []string{""},
		WorkingDirectory: "/tmp",
		StdOutPath:       "null",
		RunAsUser:        "user",
	})
}

func CreateRemoteitService() service.SystemService {
	return service.New(service.Command{
		Name:             "it.remote.cli",
		DisplayLabel:     "it.remote.cli",
		Description:      "it.remote.cli",
		DocumentationURL: "",
		Executable:       "/Users/nicolae/Downloads/remoteit_mac-osx_x86_64",
		Args:             []string{"watch", "-v", "-c", "/etc/remoteit/config.json"},
		WorkingDirectory: "",
		StdOutPath:       "null",
		RunAsUser:        "user",
	})
}
