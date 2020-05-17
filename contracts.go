package service

import (
	"github.com/codemodify/systemkit-service/spec"
)

// Info -
type Info struct {
	Error       error        `json:"-"`
	Service     spec.SERVICE `json:"config,omitempty"`
	IsRunning   bool         `json:"isRunning"`
	PID         int          `json:"pid,omitempty"`
	FilePath    string       `json:"filePath,omitempty"`
	FileContent string       `json:"fileContent,omitempty"`
}
