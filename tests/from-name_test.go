package tests

import (
	"fmt"
	"testing"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
	helpersGuid "github.com/codemodify/systemkit-helpers-guid"

	service "github.com/codemodify/systemkit-service"
)

func Test_from_name(t *testing.T) {
	service, err := service.NewServiceFromName("systemkit-test-service")
	if err != nil {
		t.Fatalf(err.Error())
	}

	info := service.Info()
	if info.Error != nil {
		fmt.Println(info.Error.Error())
	}

	fmt.Println(helpersJSON.AsJSONStringWithIndentation(info))
}

func Test_fraom_name_non_existing(t *testing.T) {
	service, err := service.NewServiceFromName(helpersGuid.NewGUID())
	if err != nil {
		t.Fatalf(err.Error())
	}

	info := service.Info()
	if info.Error != nil {
		fmt.Println(info.Error.Error())
	}

	fmt.Println(helpersJSON.AsJSONStringWithIndentation(info))
}
