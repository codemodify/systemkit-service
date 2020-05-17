package encoders

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

var logTagRC_D = "rc.d-SERVICE"

// SERVICEToUpStart -
func SERVICEToUpStart(serviceSpec spec.SERVICE) (platformService string) {
	// build export ONE='one'
	environmentAsSB := strings.Builder{}
	for key, val := range serviceSpec.Environment {
		environmentAsSB.WriteString(fmt.Sprintf("export %s=%s ", key, val))
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

	sb.WriteString("#!/bin/sh\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("# PROVIDE: %s\n", serviceSpec.Name))
	sb.WriteString(fmt.Sprintf("# REQUIRE: %s\n", dependsOnAsSB.String()))
	sb.WriteString(fmt.Sprintf("# Description: %s\n", serviceSpec.Description))
	sb.WriteString(fmt.Sprintf("# Documentation: %s\n", serviceSpec.Documentation))
	sb.WriteString("\n")
	sb.WriteString(". /etc/rc.subr")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("name=%s\n", serviceSpec.Name))
	sb.WriteString(fmt.Sprintf("rcvar=%s_enable\n", serviceSpec.Name))
	sb.WriteString("\n")
	sb.WriteString(environmentAsSB.String())

	if len(strings.TrimSpace(serviceSpec.Credentials.User)) > 0 {
		sb.WriteString(fmt.Sprintf("%s_user=\"%s\"\n", serviceSpec.Name, serviceSpec.Credentials.User))
	}
	if len(strings.TrimSpace(serviceSpec.Credentials.Group)) > 0 {
		sb.WriteString(fmt.Sprintf("%s_group=\"%s\"\n", serviceSpec.Name, serviceSpec.Credentials.Group))
	}

	sb.WriteString(fmt.Sprintf("command=\"%s\"\n", serviceSpec.Executable))
	sb.WriteString(fmt.Sprintf("command_args=\"%s\"\n", strings.Join(serviceSpec.Args, " ")))
	sb.WriteString("pidfile=\"/var/run/${name}.pid\"\n")

	if len(strings.TrimSpace(serviceSpec.WorkingDirectory)) > 0 {
		sb.WriteString(fmt.Sprintf("%s_chdir=\"%s\"\n", serviceSpec.Name, serviceSpec.WorkingDirectory))
	}

	sb.WriteString("\n")
	sb.WriteString("load_rc_config $name\n")
	sb.WriteString("run_rc_command \"$1\"\n")

	// start at boot
	rcConf, err := ioutil.ReadFile("/etc/rc.conf")
	if err == nil {
		startAtBootLineWasFound := false
		sb := strings.Builder{}
		for _, rcConfLine := range strings.Split(string(rcConf), "\n") {
			if strings.Contains(rcConfLine, serviceSpec.Name) {
				if serviceSpec.Start.AtBoot {
					sb.WriteString(serviceSpec.Name + "_enable=\"YES\"")
				}

				startAtBootLineWasFound = true
			} else {
				sb.WriteString(rcConfLine + "\n")
			}
		}

		if !startAtBootLineWasFound && serviceSpec.Start.AtBoot {
			sb.WriteString(serviceSpec.Name + "_enable=\"YES\"")
		}

		ioutil.WriteFile("/etc/rc.conf", []byte(sb.String()), 644)
	}

	return sb.String()
}
