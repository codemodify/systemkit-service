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
		Label             *string  `plist:"Label"`
		ProgramArguments  []string `plist:"ProgramArguments"`
		StandardOutPath   *string  `plist:"StandardOutPath"`
		StandardErrorPath *string  `plist:"StandardErrorPath"`
		KeepAlive         *bool    `plist:"KeepAlive"`
		WorkingDirectory  *string  `plist:"WorkingDirectory"`
	}

	if err := plist.NewXMLDecoder(strings.NewReader(platformService)).Decode(&plistData); err != nil {
		logging.Errorf("%s: error parsing PLIST: %s, from %s", logTagLaunchD, err.Error(), helpersReflect.GetThisFuncName())
	} else {
		if len(plistData.ProgramArguments) > 0 {
			serviceSpec.Executable = plistData.ProgramArguments[0]
			serviceSpec.Args = plistData.ProgramArguments[1:]
		}
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
		serviceSpec.Start.Restart = *plistData.KeepAlive
		serviceSpec.WorkingDirectory = *plistData.WorkingDirectory
	}

	return
}
