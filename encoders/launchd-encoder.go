package encoders

import (
	"fmt"
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

/*
	Based on
	https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
	https://en.wikipedia.org/wiki/Launchd
*/

var logTagLaunchD = "LaunchD-SERVICE"

// SERVICEToLaunchD -
func SERVICEToLaunchD(serviceSpec spec.SERVICE) (platformService string) {
	sb := strings.Builder{}

	// START
	sb.WriteString(fmt.Sprintf("<?xml version='1.0' encoding='UTF-8'?>\n"))
	sb.WriteString(fmt.Sprintf("<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n"))
	sb.WriteString(fmt.Sprintf("<plist version='1.0'>\n"))
	sb.WriteString(fmt.Sprintf("	<dict>\n"))

	// SERVICE.Name
	sb.WriteString(fmt.Sprintf("		<key>Label</key>\n"))
	sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", serviceSpec.Name))
	sb.WriteString("\n")

	// SERVICE.Description
	// FIXME: can't find the doc how to add

	// SERVICE.Documentation
	// FIXME: can't find the doc how to add

	// SERVICE.Executable + SERVICE.Args
	args := []string{serviceSpec.Executable}
	if len(serviceSpec.Args) > 0 {
		args = append(args, serviceSpec.Args...)
	}
	sb.WriteString(fmt.Sprintf("		<key>ProgramArguments</key>\n"))
	sb.WriteString(fmt.Sprintf("		<array>\n"))
	for _, arg := range args {
		sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", arg))
	}
	sb.WriteString(fmt.Sprintf("		</array>\n"))

	// SERVICE.WorkingDirectory
	if len(strings.TrimSpace(serviceSpec.WorkingDirectory)) > 0 {
		sb.WriteString(fmt.Sprintf("		<key>WorkingDirectory</key>\n"))
		sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", serviceSpec.WorkingDirectory))
	}

	// SERVICE.Environment
	// FIXME: can't find the doc how to add

	// SERVICE.DependsOn
	// FIXME: can't find the doc how to add

	// SERVICE.Start
	if serviceSpec.Start.AtBoot {
		sb.WriteString(fmt.Sprintf("		<key>RunAtLoad</key>\n"))
		sb.WriteString(fmt.Sprintf("		<true />\n"))
	}
	if serviceSpec.Start.Restart {
		sb.WriteString(fmt.Sprintf("		<key>KeepAlive</key>\n"))
		sb.WriteString(fmt.Sprintf("		<true />\n"))
	}

	// SERVICE.Logging
	if !serviceSpec.Logging.StdOut.Disabled && !serviceSpec.Logging.StdOut.UseDefault {
		sb.WriteString(fmt.Sprintf("		<key>StandardOutPath</key>\n"))
		sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", serviceSpec.Logging.StdOut.Value))
	}
	if !serviceSpec.Logging.StdErr.Disabled && !serviceSpec.Logging.StdErr.UseDefault {
		sb.WriteString(fmt.Sprintf("		<key>StandardErrorPath</key>\n"))
		sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", serviceSpec.Logging.StdErr.Value))
	}

	// SERVICE.Credentials
	if len(strings.TrimSpace(serviceSpec.Credentials.User)) > 0 {
		sb.WriteString(fmt.Sprintf("		<key>UserName</key>\n"))
		sb.WriteString(fmt.Sprintf("		<string>%s</string>\n", serviceSpec.Credentials.User))
	}
	// FIXME: can't find the doc how to add SERVICE.Credentials.Group

	// END
	sb.WriteString(fmt.Sprintf("	</dict>\n"))
	sb.WriteString(fmt.Sprintf("</plist>\n"))

	return sb.String()
}
