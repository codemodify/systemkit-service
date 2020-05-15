package main

import (
	"os"

	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	service "github.com/codemodify/systemkit-service"
)

type infoCommandFlags struct {
	Name string `flagName:"name" flagRequired:"true" flagDescription:"Service name"`
}

func init() {
	rootCommand.AddCommand(&clicmdflags.Command{
		Name:        "info",
		Description: "Query a system service",
		Examples: []string{
			"-name test-service",
		},
		Flags: infoCommandFlags{},
		Handler: func(command *clicmdflags.Command) {
			opStatus := OperationStatus{
				Status:  OpStatusSuccess,
				Details: []string{},
			}

			flags, ok := command.Flags.(infoCommandFlags)
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

			info := s.Info()
			info.FileContent = ""

			os.Stdout.WriteString(helpersJSON.AsJSONStringWithIndentation(info) + "\n")
		},
	})
}
