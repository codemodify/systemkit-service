package main

import (
	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
	service "github.com/codemodify/systemkit-service"
)

type deleteCommandFlags struct {
	Name string `flagName:"name" flagRequired:"true" flagDescription:"Service name"`
}

func init() {
	rootCommand.AddCommand(&clicmdflags.Command{
		Name:        "delete",
		Description: "Delete a system service",
		Examples: []string{
			"-name test-service",
		},
		Flags: deleteCommandFlags{},
		Handler: func(command *clicmdflags.Command) {
			opStatus := OperationStatus{
				Status:  OpStatusSuccess,
				Details: []string{},
			}

			flags, ok := command.Flags.(deleteCommandFlags)
			if !ok {
				opStatus.Status = OpStatusError
				opStatus.Details = append(opStatus.Details, "Can't fetch flags values")
				logOpearationStatus(opStatus)
				return
			}

			s, err := service.NewServiceFromName(flags.Name)
			if err != nil {
				opStatus.Status = OpStatusError
				opStatus.Details = append(opStatus.Details, err.Error())
				logOpearationStatus(opStatus)
				return
			}

			err = s.Uninstall()
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
