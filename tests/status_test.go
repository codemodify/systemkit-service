package tests

import (
	"fmt"
	"testing"
)

func Test_status(t *testing.T) {
	service := createService()

	serviceStatus := service.Status()
	if serviceStatus.Error != nil {
		t.Fatalf(serviceStatus.Error.Error())
	}

	fmt.Println(serviceStatus)
}

func Test_status_non_existing(t *testing.T) {
	service := createRandomService()

	serviceStatus := service.Status()
	if serviceStatus.Error != nil {
		t.Fatalf(serviceStatus.Error.Error())
	}

	fmt.Println(serviceStatus)
}
