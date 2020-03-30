# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) Service
[![GoDoc](https://godoc.org/github.com/codemodify/systemkit-logging?status.svg)](https://godoc.org/github.com/codemodify/systemkit-events)
[![0-License](https://img.shields.io/badge/license-0--license-brightgreen)](https://github.com/codemodify/TheFreeLicense)
[![Go Report Card](https://goreportcard.com/badge/github.com/codemodify/systemkit-logging)](https://goreportcard.com/report/github.com/codemodify/systemkit-logging)
[![Test Status](https://github.com/danawoodman/systemservice/workflows/Test/badge.svg)](https://github.com/danawoodman/systemservice/actions)
![code size](https://img.shields.io/github/languages/code-size/codemodify/SystemKit?style=flat-square)

#### Robust Cross platform Create/Start/Stop/Delete system or user service.
#### Supported: Linux, Raspberry Pi, FreeBSD, Mac OS, Windows, Solaris

# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) Install
```go
go get github.com/codemodify/systemkit-service
```

# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) API

&nbsp;										| &nbsp;
---     									| ---
service := Service.New()                    | Create a new system service
service.Install(false)	                    | Install a new system service
service.Install(true)                       | Install a new system service and start
service.Start()                             | Start system service 
service.Restart()                           | Restart system service 
service.Status()                            | System service status
service.Stop()                              | Stop system service 
service.Uninstall()                         | Uninstall system service 


# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) Usage
```go
package main

import (
	"fmt"

	"https://github.com/codemodify/systemkit-service/"
)

func main() {

// Create service definition

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

// Instal and start
	
	err := service.Install(true)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Done
```

---

> `IMPORTANT NOTE:`<br>If this is executed as `SUDO` then a system service will be created instead of user service
