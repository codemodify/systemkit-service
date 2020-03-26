# ![](https://fonts.gstatic.com/s/i/materialicons/label_important/v4/24px.svg) Service
[![GoDoc](https://godoc.org/github.com/codemodify/SystemKit?status.svg)](https://godoc.org/github.com/codemodify/SystemKit)
[![0-License](https://img.shields.io/badge/license-0--license-brightgreen)](https://github.com/codemodify/TheFreeLicense)
[![Go Report Card](https://goreportcard.com/badge/github.com/codemodify/SystemKit)](https://goreportcard.com/report/github.com/codemodify/SystemKit)
[![Test Status](https://github.com/danawoodman/systemservice/workflows/Test/badge.svg)](https://github.com/danawoodman/systemservice/actions)
![code size](https://img.shields.io/github/languages/code-size/codemodify/SystemKit?style=flat-square)

Cross platform Create/Start/Stop/Delete system or user service
	- How to use?	
```go
package main

import (
	"fmt"

	"github.com/codemodify/SystemKit/Service"
)

func main() {
	// 1. Download a sample service definiton from here: `github.com/codemodify/SystemKit/Service/samples/MacOS/service.sh` to your `~/Downloads` folder

	// 2. Create service definition
	usr, _ := user.Current()
	service := Service.New(Service.Command{
		Name:             "MY_SERVICE",
		DisplayLabel:     "My Service",
		Description:      "This service is a test service",
		DocumentationURL: "",
		Executable:       usr.HomeDir + "/Downloads/service.sh"),
		Args:             []string{""},
		WorkingDirectory: usr.HomeDir,
	})

	// 3. Instal and start
	err := service.Install(true)
	if err != nil {
		fmt.Println(err.Error())
	}

	// 4. DONE
	// NOTE: if this is executed as SUDO then a system service will be created instead of user service

}

```