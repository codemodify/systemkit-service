// +build !windows

package main

import (
	"fmt"

	"github.com/codemodify/systemkit-service/tests"
)

func main() {
	service := tests.CreateRemoteitService()

	// err := service.Stop()
	// if helpersErrors.Is(err, Service.ErrServiceDoesNotExist) {
	// 	// this is a good thing
	// } else if err != nil {
	// 	fmt.Println(err.Error())
	// }

	err := service.Uninstall()
	if err != nil {
		fmt.Println(err.Error())
	}
}
