package tests

import (
	"fmt"
	"testing"
)

func Test_status(t *testing.T) {
	service := createService()

	serviceInfo := service.Info()
	if serviceInfo.Error != nil {
		t.Fatalf(serviceInfo.Error.Error())
	}

	fmt.Println(serviceInfo)
}

func Test_status_non_existing(t *testing.T) {
	service := createRandomService()

	serviceInfo := service.Info()
	if serviceInfo.Error != nil {
		t.Fatalf(serviceInfo.Error.Error())
	}

	fmt.Println(serviceInfo)
}
