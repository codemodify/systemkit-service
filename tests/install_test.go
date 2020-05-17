package tests

import (
	"testing"
)

func Test_install(t *testing.T) {
	service := CreateService()

	err := service.Install()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_install_and_start(t *testing.T) {
	service := CreateService()

	err := service.Install()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
