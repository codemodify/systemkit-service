package tests

import (
	"testing"
)

func Test_install(t *testing.T) {
	service := createService()

	err := service.Install(false)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_install_and_start(t *testing.T) {
	service := createService()

	err := service.Install(true)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
