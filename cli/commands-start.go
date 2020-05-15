package main

import (
	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
	service "github.com/codemodify/systemkit-service"
)

type startCommandFlags struct {
	Name string `flagName:"name" flagRequired:"true" flagDescription:"Service name"`
}

func init() {
	rootCommand.AddCommand(&clicmdflags.Command{
		Name:        "start",
		Description: "Start a system service",
		Examples: []string{
			"-name test-service",
		},
		Flags: startCommandFlags{},
		Handler: func(command *clicmdflags.Command) {
			opStatus := OperationStatus{
				Status:  OpStatusSuccess,
				Details: []string{},
			}

			flags, ok := command.Flags.(startCommandFlags)
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

			err = s.Start()
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
