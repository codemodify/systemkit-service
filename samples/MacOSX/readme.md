# ![](https://fonts.gstatic.com/s/i/materialicons/label_important/v4/24px.svg) Service
[![GoDoc](https://godoc.org/github.com/codemodify/SystemKit?status.svg)](https://godoc.org/github.com/codemodify/SystemKit)
[![0-License](https://img.shields.io/badge/license-0--license-brightgreen)](https://github.com/codemodify/TheFreeLicense)
[![Go Report Card](https://goreportcard.com/badge/github.com/codemodify/SystemKit)](https://goreportcard.com/report/github.com/codemodify/SystemKit)
[![Test Status](https://github.com/danawoodman/systemservice/workflows/Test/badge.svg)](https://github.com/danawoodman/systemservice/actions)
![code size](https://img.shields.io/github/languages/code-size/codemodify/SystemKit?style=flat-square)

# Cross platform <h3> `CREATE/START/STOP/UNINSTALL` SYSTEM / USER SERVICE
## How to use?

# ![](https://fonts.gstatic.com/s/i/materialicons/label_important/v4/24px.svg) MacOS ![](https://img.icons8.com/ios-filled/30/000000/mac-os.png)
 - 	Download a sample service definiton from here and save/extract to your `~/Downloads` folder:

> `https://github.com/codemodify/systemkit-service/samples/service.sh`

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
#### `INSTALL WITHOUT START  `
```go 
	err := service.Install(false)
	if err != nil {
		fmt.Println(err.Error())
	
```
---
#### `STOP SERVICE  `
```go 
	err := service.Stop
	if err != nil {
		fmt.Println(err.Error())
```
#### `UNINSTALL SERVICE  `
```go 
	err := service.Uninstall
	if err != nil {
		fmt.Println(err.Error())
```
---
#### `IMPORTANT NOTE` 
### If this is executed as `SUDO` then a system service will be created instead of user service
