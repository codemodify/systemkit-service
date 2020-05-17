package encoders

import (
	"strconv"
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

// SystemDToSERVICE -
func SystemDToSERVICE(platformService string) (serviceSpec spec.SERVICE) {
	serviceSpec = newEmptySERVICE()

	for _, line := range strings.Split(platformService, "\n") {

		// [Unit]
		if strings.Contains(line, "Description=") {
			serviceSpec.Description = strings.TrimSpace(strings.Replace(line, "Description=", "", 1))

		} else if strings.Contains(line, "Documentation=") {
			serviceSpec.Documentation = strings.TrimSpace(strings.Replace(line, "Documentation=", "", 1))

		} else if strings.Contains(line, "Requires=") || strings.Contains(line, "After=") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "Requires=", "", 1))
			cleanLine = strings.TrimSpace(strings.Replace(line, "After=", "", 1))
			for _, item := range strings.Split(cleanLine, " ") {
				serviceSpec.DependsOn = append(serviceSpec.DependsOn, spec.ServiceType(item))
			}

		} else if strings.Contains(line, "StartLimitIntervalSec=") {
			serviceSpec.Start.Restart = true
			cleanLine := strings.TrimSpace(strings.Replace(line, "StartLimitIntervalSec=", "", 1))
			serviceSpec.Start.RestartTimeout, _ = strconv.Atoi(cleanLine)

			// [Service]
		} else if strings.Contains(line, "ExecStart=") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "ExecStart=", "", 1))
			parts := strings.Split(cleanLine, " ")
			serviceSpec.Executable = parts[0]
			serviceSpec.Args = parts[1:]

		} else if strings.Contains(line, "WorkingDirectory=") {
			serviceSpec.WorkingDirectory = strings.TrimSpace(strings.Replace(line, "WorkingDirectory=", "", 1))

		} else if strings.Contains(line, "Environment=") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "Environment=", "", 1))
			keyValues := strings.Split(cleanLine, "=")
			for i := 0; i < len(keyValues); i++ {
				key := ""
				val := ""

				if i%2 == 0 {
					key = keyValues[i]

					if (i + 1) <= (len(keyValues) - 1) {
						val = keyValues[i+1]
					}

					serviceSpec.Environment[key] = val
				}
			}

		} else if strings.Contains(line, "Restart=") {
			serviceSpec.Start.Restart = true

		} else if strings.Contains(line, "RestartSec=") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "RestartSec=", "", 1))
			serviceSpec.Start.RestartTimeout, _ = strconv.Atoi(cleanLine)

		} else if strings.Contains(line, "StandardOutput=") {
			serviceSpec.Logging.StdOut.Disabled = false
			serviceSpec.Logging.StdOut.UseDefault = false
			serviceSpec.Logging.StdOut.Value = strings.TrimSpace(strings.Replace(line, "StandardOutput=", "", 1))

		} else if strings.Contains(line, "StandardError=") {
			serviceSpec.Logging.StdErr.Disabled = false
			serviceSpec.Logging.StdErr.UseDefault = false
			serviceSpec.Logging.StdErr.Value = strings.TrimSpace(strings.Replace(line, "StandardError=", "", 1))

		} else if strings.Contains(line, "User=") {
			serviceSpec.Credentials.User = strings.TrimSpace(strings.Replace(line, "User=", "", 1))

		} else if strings.Contains(line, "Group=") {
			serviceSpec.Credentials.Group = strings.TrimSpace(strings.Replace(line, "Group=", "", 1))
		}
	}

	return
}
