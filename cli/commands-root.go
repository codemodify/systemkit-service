package main

import (
	"os"
	"path/filepath"

	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
)

type rootCommandFlags struct {
	JSON    bool `flagName:"json"    flagDefault:"false" flagDescription:"Enables JSON output"`
	Verbose bool `flagName:"verbose" flagDefault:"false" flagDescription:"Enables verbose output"`
}

var rootCommand = &clicmdflags.Command{
	Name:        filepath.Base(os.Args[0]),
	Description: "Create/Remove/Start/Stop/Query a system service",
	Examples: []string{
		filepath.Base(os.Args[0]) + " -json",
		filepath.Base(os.Args[0]) + " -json true",
	},
	Flags: rootCommandFlags{},
}

func globalFlags() rootCommandFlags {
	flags, ok := rootCommand.Flags.(rootCommandFlags)
	if !ok {
		return rootCommandFlags{}
	}

	return flags
}
