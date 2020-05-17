package tests

import (
	"fmt"
	"testing"

	helpersJSON "github.com/codemodify/systemkit-helpers-conv"
)

func Test_status(t *testing.T) {
	service := CreateService()

	info := service.Info()
	if info.Error != nil {
		fmt.Println(info.Error.Error())
	}

	fmt.Println(helpersJSON.AsJSONStringWithIndentation(info))
}

func Test_status_non_existing(t *testing.T) {
	service := CreateRandomService()

	info := service.Info()
	if info.Error != nil {
		fmt.Println(info.Error.Error())
	}

	fmt.Println(helpersJSON.AsJSONStringWithIndentation(info))
}
