package main

import (
	"strings"

	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
	service "github.com/codemodify/systemkit-service"
)

type createCommandFlags struct {
	Name        string `flagName:"name"        flagRequired:"true"  flagDescription:"Service name"`
	Description string `flagName:"description" flagRequired:"false" flagDescription:"Service description"`
	Executable  string `flagName:"executable"  flagRequired:"true"  flagDescription:"Service executable"`
	Args        string `flagName:"args"        flagRequired:"false" flagDescription:"Executable args"`
}

func init() {
	rootCommand.AddCommand(&clicmdflags.Command{
		Name:        "create",
		Description: "Create a system service",
		Examples: []string{
			"-name test-service -executable htop",
		},
		Flags: createCommandFlags{},
		Handler: func(command *clicmdflags.Command) {
			flags := command.Flags.(createCommandFlags)

			s := service.NewServiceFromConfig(service.Config{
				Name:        flags.Name,
				Description: flags.Description,
				Executable:  flags.Executable,
				Args:        strings.Split(flags.Args, " "),
			})

			err := s.Install()
		},
	})
}
