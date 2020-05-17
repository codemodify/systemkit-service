package tests

import (
	"testing"
)

func Test_start(t *testing.T) {
	service := CreateService()

	err := service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_start_non_existing(t *testing.T) {
	service := CreateRandomService()

	err := service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
