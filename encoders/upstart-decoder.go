package encoders

import (
	"strings"

	"github.com/codemodify/systemkit-service/spec"
)

// UpStartToSERVICE -
func UpStartToSERVICE(platformService string) (serviceSpec spec.SERVICE) {
	serviceSpec = newEmptySERVICE()

	for lineIndex, line := range strings.Split(platformService, "\n") {
		if strings.Contains(line, "# ") && lineIndex == 0 {
			serviceSpec.Description = strings.TrimSpace(strings.Replace(line, "# ", "", 1))

		} else if strings.Contains(line, "exec ") {
			cleanLine := strings.TrimSpace(strings.Replace(line, "exec ", "", 1))
			parts := strings.Split(cleanLine, " ")
			serviceSpec.Executable = parts[0]
			serviceSpec.Args = parts[1:]

		}
	}

	return
}
