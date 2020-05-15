package main

import (
	"os"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
)

type OpStatusType int

const (
	OpStatusError   OpStatusType = -1
	OpStatusSuccess              = 0
	OpStatusWarning              = 1
)

type OperationStatus struct {
	Status  OpStatusType `json:"status"`
	Details []string     `json:"details,omitempty"`
}

func logOpearationStatus(opStatus OperationStatus) {
	if globalFlags().JSON {
		os.Stdout.WriteString(helpersJSON.AsJSONString(opStatus))
	} else {
		if opStatus.Status == OpStatusError {
			for _, v := range opStatus.Details {
				os.Stderr.WriteString(v + "\n")
			}
		} else if globalFlags().Verbose {
			for _, v := range opStatus.Details {
				os.Stdout.WriteString(v + "\n")
			}
		}
	}
}
