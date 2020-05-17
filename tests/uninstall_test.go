package tests

import (
	"testing"
)

func Test_uninstall(t *testing.T) {
	service := CreateService()

	err := service.Uninstall()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_uninstall_non_existing(t *testing.T) {
	service := CreateRandomService()

	err := service.Uninstall()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
