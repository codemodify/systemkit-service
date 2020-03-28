// +build darwin

package main

import (
	"fmt"
	"os/user"

	"github.com/codemodify/SystemKit/Service"
)

func main() {
	usr, _ := user.Current()

	service := Service.New(Service.Command{
		Name:             "MY_SERVICE",
		DisplayLabel:     "My Service",
		Description:      "This service is a test service",
		DocumentationURL: "",
		Executable:       usr.HomeDir + "/Downloads/service.sh",
		Args:             []string{""},
		WorkingDirectory: usr.HomeDir,
	})

	err := service.Install(true)
	if err != nil {
		fmt.Println(err.Error())
	}
}
