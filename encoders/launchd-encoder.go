package encoders

import (
	"bytes"
	"text/template"

	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
	"github.com/codemodify/systemkit-service/spec"
)

var logTagLaunchD = "LaunchD-SERVICE"

func SERVICEToLaunchD(serviceSpec spec.SERVICE) (platformService string) {
	// for LaunchD move everything into serviceSpec.Args
	args := []string{serviceSpec.Executable}

	if len(serviceSpec.Args) > 0 {
		args = append(args, serviceSpec.Args...)
	}

	serviceSpec.Args = args

	// run the template
	fileTemplate := template.Must(template.New("launchdFile").Parse(`
<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
<plist version='1.0'>
	<dict>
		<key>Label</key>
		<string>{{.Name}}</string>

		<key>ProgramArguments</key>
		<array>{{ range $arg := .Args}}
			<string>{{ $arg}}</string>{{ end}}
		</array>

		{{ if eq .StdOut.Disable false}}
		<key>StandardOutPath</key>
		<string>{{.StdOut.Value}}</string>
		{{ end}}

		{{ if eq .StdErr.Disable false}}
		<key>StandardErrorPath</key>
		<string>{{.StdErr.Value}}</string>
		{{ end}}

		<key>KeepAlive</key>
		<{{.Restart}}/>
		<key>RunAtLoad</key>
		<true/>

		<key>WorkingDirectory</key>
		<string>{{.WorkingDirectory}}</string>
	</dict>
</plist>
`))

	var buffer bytes.Buffer
	if err := fileTemplate.Execute(&buffer, serviceSpec); err != nil {
		logging.Errorf("%s: error generating file: %s, from %s", logTagLaunchD, err.Error(), helpersReflect.GetThisFuncName())
		return ""
	}

	return buffer.String()
}
