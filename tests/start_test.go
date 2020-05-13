package tests

import (
	"testing"
)

func Test_start(t *testing.T) {
	service := createService()

	err := service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_start_non_existing(t *testing.T) {
	service := createRandomService()

	err := service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
