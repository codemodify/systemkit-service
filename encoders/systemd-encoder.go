package encoders

import (
	"fmt"
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

/*
Based on https://www.freedesktop.org/software/systemd/man/systemd.unit.html we work with the following
- System Unit Search Path
	- /etc/systemd/system/*
- User Unit Search Path
	- ~/.config/systemd/user/*
- [Unit]
	- Description=
	- Documentation=
	- Requires=
	- StartLimitIntervalSec=
	- StartLimitBurst=0
	- StartLimitAction=none
- [Install]
	- WantedBy=multi-user.target

Based on https://www.freedesktop.org/software/systemd/man/systemd.service.html we work with the following
- [Service]
	- Type=simple
	- ExecStart=
	- Restart= always / on-failure
	- RestartSec=
	- StandardOutput=
	- StandardError=
	- User=
	- Group=
*/

// SERVICEToSystemD -
func SERVICEToSystemD(serviceSpec spec.SERVICE) (platformService string) {
	// for SystemD move everything into serviceSpec.Executable
	if len(serviceSpec.Args) > 0 {
		serviceSpec.Executable = fmt.Sprintf(
			"%s %s",
			serviceSpec.Executable,
			strings.Join(serviceSpec.Args, " "),
		)
	}

	// build `Environment=`, ex: Environment=ONE='one' "TWO='two two' too" THREE=
	environmentAsSB := strings.Builder{}
	for key, val := range serviceSpec.Environment {
		environmentAsSB.WriteString(fmt.Sprintf("%s=%s ", key, val))
	}

	// build the `Requires=`
	// FIXME: parse the DependsOnOverrides
	dependsOnAsSB := strings.Builder{}
	for _, dependsOn := range serviceSpec.DependsOn {
		if systemdServices, ok := serviceMappings[spec.InitSystemd]; ok { // check if a key is in dictionary
			if systemdService, ok := systemdServices[spec.ServiceType(dependsOn)]; ok {
				dependsOnAsSB.WriteString(systemdService + " ")
			}
		}
	}

	sb := strings.Builder{}

	// [Unit]
	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s\n", serviceSpec.Description))
	sb.WriteString(fmt.Sprintf("Documentation=%s\n", serviceSpec.Documentation))
	sb.WriteString(fmt.Sprintf("Requires=%s\n", dependsOnAsSB.String()))
	if serviceSpec.Start.Restart {
		sb.WriteString(fmt.Sprintf("StartLimitIntervalSec=%d\n", serviceSpec.Start.RestartTimeout))
		sb.WriteString(fmt.Sprintf("StartLimitBurst=0\n"))
		sb.WriteString(fmt.Sprintf("StartLimitAction=none\n"))
	}

	// [Service]
	sb.WriteString("[Service]\n")
	sb.WriteString("Type=simple\n")
	sb.WriteString(fmt.Sprintf("ExecStart=\n", serviceSpec.Executable))

	if len(strings.TrimSpace(serviceSpec.WorkingDirectory)) > 0 {
		sb.WriteString(fmt.Sprintf("WorkingDirectory=%s\n", serviceSpec.WorkingDirectory))
	}

	if environmentAsSB.Len() > 0 {
		sb.WriteString(fmt.Sprintf("Environment=%s\n", environmentAsSB.String()))
	}

	if serviceSpec.Start.Restart {
		sb.WriteString(fmt.Sprintf("Restart=always\n"))
		sb.WriteString(fmt.Sprintf("RestartSec=%d\n", serviceSpec.Start.RestartTimeout))
	} else {
		sb.WriteString(fmt.Sprintf("Restart=on-failure\n"))
	}

	if serviceSpec.Logging.StdOut.Disabled {
		sb.WriteString("StandardOutput=null\n")
	} else if !serviceSpec.Logging.StdOut.UseDefault {
		sb.WriteString(fmt.Sprintf("StandardOutput=%s\n", serviceSpec.Logging.StdOut.Value))
	}
	if serviceSpec.Logging.StdErr.Disabled {
		sb.WriteString("StandardOutput=null\n")
	} else if !serviceSpec.Logging.StdErr.UseDefault {
		sb.WriteString(fmt.Sprintf("StandardError=%s\n", serviceSpec.Logging.StdErr.Value))
	}

	if len(strings.TrimSpace(serviceSpec.Credentials.User)) > 0 {
		sb.WriteString(fmt.Sprintf("User=\n", serviceSpec.Credentials.User))
	}
	if len(strings.TrimSpace(serviceSpec.Credentials.Group)) > 0 {
		sb.WriteString(fmt.Sprintf("Group=\n", serviceSpec.Credentials.Group))
	}

	// [Install]
	if serviceSpec.Start.AtBoot {
		sb.WriteString("[Install]\n")
		sb.WriteString("WantedBy=multi-user.target\n")
	}

	return sb.String()
}
