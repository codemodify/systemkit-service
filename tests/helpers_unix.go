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
