package encoders

import (
	"strings"

	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
	"github.com/codemodify/systemkit-service/spec"
	"github.com/groob/plist"
)

// LaunchDToSERVICE -
func LaunchDToSERVICE(platformService string) (serviceSpec spec.SERVICE) {
	serviceSpec = newEmptySERVICE()

	var plistData struct {
		Label             *string  `plist:"Label"`             // SERVICE.Name
		ProgramArguments  []string `plist:"ProgramArguments"`  // SERVICE.Executable + SERVICE.Args
		WorkingDirectory  *string  `plist:"WorkingDirectory"`  // SERVICE.WorkingDirectory
		RunAtLoad         *bool    `plist:"RunAtLoad"`         // SERVICE.Start
		KeepAlive         *bool    `plist:"KeepAlive"`         // SERVICE.Start
		StandardOutPath   *string  `plist:"StandardOutPath"`   // SERVICE.Logging
		StandardErrorPath *string  `plist:"StandardErrorPath"` // SERVICE.Logging
		UserName          *string  `plist:"UserName"`          // SERVICE.Credentials
	}

	if err := plist.NewXMLDecoder(strings.NewReader(platformService)).Decode(&plistData); err != nil {
		logging.Errorf("%s: error parsing PLIST: %s, from %s", logTagLaunchD, err.Error(), helpersReflect.GetThisFuncName())
	} else {
		// SERVICE.Name
		if plistData.Label != nil {
			serviceSpec.Name = *plistData.Label
		}

		// SERVICE.Executable + SERVICE.Args
		if len(plistData.ProgramArguments) > 0 {
			serviceSpec.Executable = plistData.ProgramArguments[0]

			if len(plistData.ProgramArguments) > 1 {
				serviceSpec.Args = plistData.ProgramArguments[1:]
			}
		}

		// SERVICE.WorkingDirectory
		if plistData.WorkingDirectory != nil {
			serviceSpec.WorkingDirectory = *plistData.WorkingDirectory
		}

		// SERVICE.Start
		if plistData.RunAtLoad != nil {
			serviceSpec.Start.AtBoot = *plistData.RunAtLoad
		}
		if plistData.KeepAlive != nil {
			serviceSpec.Start.Restart = *plistData.KeepAlive
		}

		// SERVICE.Logging
		if plistData.StandardOutPath != nil {
			serviceSpec.Logging.StdOut.Disabled = false
			serviceSpec.Logging.StdOut.UseDefault = false
			serviceSpec.Logging.StdOut.Value = *plistData.StandardOutPath
		}
		if plistData.StandardErrorPath != nil {
			serviceSpec.Logging.StdErr.Disabled = false
			serviceSpec.Logging.StdErr.UseDefault = false
			serviceSpec.Logging.StdErr.Value = *plistData.StandardErrorPath
		}

		// SERVICE.Credentials
		if plistData.UserName != nil {
			serviceSpec.Credentials.User = *plistData.UserName
		}
	}

	return
}
