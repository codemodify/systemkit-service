package encoders

import (
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

// SystemVToSERVICE -
func SystemVToSERVICE(platformService string) (serviceSpec spec.SERVICE) {
	serviceSpec = newEmptySERVICE()

	for _, line := range strings.Split(platformService, "\n") {
		if strings.Contains(line, "# Description:") {
			serviceSpec.Description = strings.TrimSpace(strings.Replace(line, "# Description:", "", 1))

		} else if strings.Contains(line, "cmd=\"") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "cmd=\"", "", 1))
			parts := strings.Split(cleanLine, " ")
			serviceSpec.Executable = parts[0]
			serviceSpec.Args = parts[1:]

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
