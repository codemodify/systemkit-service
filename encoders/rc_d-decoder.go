package encoders

import (
	"fmt"
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

// RC_DToSERVICE -
func RC_DToSERVICE(platformService string) (serviceSpec spec.SERVICE) {
	serviceSpec = newEmptySERVICE()

	for _, line := range strings.Split(platformService, "\n") {
		if strings.Contains(line, "# PROVIDE:") {
			serviceSpec.Name = strings.TrimSpace(strings.Replace(line, "# PROVIDE:", "", 1))

		} else if strings.Contains(line, "# REQUIRE:") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "# REQUIRE:", "", 1))
			for _, item := range strings.Split(cleanLine, " ") {
				serviceSpec.DependsOn = append(serviceSpec.DependsOn, spec.ServiceType(item))
			}

		} else if strings.Contains(line, "# Description:") {
			serviceSpec.Description = strings.TrimSpace(strings.Replace(line, "# Description:", "", 1))

		} else if strings.Contains(line, "# Documentation") {
			serviceSpec.Documentation = strings.TrimSpace(strings.Replace(line, "# Documentation", "", 1))

		} else if strings.Contains(line, "name=") {
			serviceSpec.Name = strings.TrimSpace(strings.Replace(line, "name=", "", 1))

		} else if strings.Contains(line, "export ") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "export ", "", 1))
			keyValue := strings.Split(cleanLine, "=")
			if len(keyValue) > 0 {
				serviceSpec.Environment[keyValue[0]] = ""
			}
			if len(keyValue) > 1 {
				serviceSpec.Environment[keyValue[0]] = keyValue[1]
			}
		} else if strings.Contains(line, fmt.Sprintf("%s_user=", serviceSpec.Name)) {
			serviceSpec.Credentials.User = strings.TrimSpace(strings.Replace(line, fmt.Sprintf("%s_user=", serviceSpec.Name), "", 1))

		} else if strings.Contains(line, fmt.Sprintf("%s_group=", serviceSpec.Name)) {
			serviceSpec.Credentials.Group = strings.TrimSpace(strings.Replace(line, fmt.Sprintf("%s_group=", serviceSpec.Name), "", 1))

		} else if strings.Contains(line, fmt.Sprintf("%s_group=", serviceSpec.Name)) {
			serviceSpec.Credentials.Group = strings.TrimSpace(strings.Replace(line, fmt.Sprintf("%s_group=", serviceSpec.Name), "", 1))

		} else if strings.Contains(line, "command=") {
			serviceSpec.Executable = strings.TrimSpace(strings.Replace(line, "command=", "", 1))

		} else if strings.Contains(line, "command_args=") {
			serviceSpec.Executable = strings.TrimSpace(strings.Replace(line, "command_args=", "", 1))

		} else if strings.Contains(line, "_chdir=") {
			serviceSpec.WorkingDirectory = strings.TrimSpace(strings.Replace(line, "_chdir=", "", 1))

		} else if strings.Contains(line, "stdout_log=\"") {
			serviceSpec.Logging.StdOut.Disabled = false
			serviceSpec.Logging.StdOut.UseDefault = false
			serviceSpec.Logging.StdOut.Value = strings.TrimSpace(strings.Replace(line, "stdout_log=\"", "", 1))

		} else if strings.Contains(line, "stderr_log=\"") {
			serviceSpec.Logging.StdErr.Disabled = false
			serviceSpec.Logging.StdErr.UseDefault = false
			serviceSpec.Logging.StdErr.Value = strings.TrimSpace(strings.Replace(line, "stderr_log=\"", "", 1))

		}
	}

	return
}
