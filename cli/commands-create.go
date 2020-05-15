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
			opStatus := OperationStatus{
				Status:  OpStatusSuccess,
				Details: []string{},
			}

			flags, ok := command.Flags.(createCommandFlags)
			if !ok {
				opStatus.Status = OpStatusError
				opStatus.Details = append(opStatus.Details, "Can't fetch flags values")
				logOpearationStatus(opStatus)
				return
			}

			s := service.NewServiceFromConfig(service.Config{
				Name:        flags.Name,
				Description: flags.Description,
				Executable:  flags.Executable,
				Args:        strings.Split(flags.Args, " "),
			})

			err := s.Install()
			if err != nil {
				opStatus.Status = OpStatusError
				opStatus.Details = append(opStatus.Details, err.Error())
				logOpearationStatus(opStatus)
				return
			}

			opStatus.Details = append(opStatus.Details, "OK")
			logOpearationStatus(opStatus)
		},
	})
}
