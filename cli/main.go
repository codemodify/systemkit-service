package main

import (
	"log"
	"os"
	"path/filepath"

	clicmdflags "github.com/codemodify/systemkit-clicmdflags"
	service "github.com/codemodify/systemkit-service"
)

type cmdFlags struct {
	Name        string `flagName:"name" 		flagRequired:"true" 	flagDescription:"Service name"`
	Description string `flagName:"description" 	flagRequired:"false" 	flagDescription:"Service description"`
	Executable  string `flagName:"executable" 	flagRequired:"true" 	flagDescription:"Service executable"`
	Args        string `flagName:"args" 		flagRequired:"false" 	flagDescription:"Executable args"`
	JSON        bool   `flagName:"json"			flagDefault:"false"		flagDescription:"Enables JSON output"`
	Verbose     bool   `flagName:"verbose" 		flagDefault:"false" 	flagDescription:"Enables verbose output"`
}

func main() {
	var cmd = &clicmdflags.Command{
		Name:        filepath.Base(os.Args[0]),
		Description: "Create a system service",
		Examples: []string{
			filepath.Base(os.Args[0]) + " -json",
			filepath.Base(os.Args[0]) + " -json true",
		},
		Flags:   cmdFlags{},
		Handler: handler,
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func handler(command *clicmdflags.Command) {
	service.NewServiceFromConfig(service.Config{
		Name:               "systemkit-test-service",
		Description:        "SystemKit Test Service",
		Executable:         "/bin/sleep",
		Args:               []string{"40"},
		WorkingDirectory:   "/tmp",
		Restart:            true,
		DelayBeforeRestart: 10,
		StdOut: service.LogConfig{
			Disable: true,
		},
		StdErr: service.LogConfig{
			Disable: true,
		},
	})
}
