package tests

import (
	"fmt"
	"testing"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	service "github.com/codemodify/systemkit-service"
)

func Test_from_name(t *testing.T) {
	service, err := service.NewServiceFromName("systemkit-test-service")
	if err != nil {
		t.Fatalf(err.Error())
	}

	info := service.Info()
	fmt.Println(helpersJSON.AsJSONStringWithIndentation(info))
}

// func Test_uninstall_non_existing(t *testing.T) {
// 	service := createRandomService()

// 	err := service.Uninstall()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// }
