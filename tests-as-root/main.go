package main

import (
	"fmt"
)

func main() {
	service := tests.createRemoteitService()

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
