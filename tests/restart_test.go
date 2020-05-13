package tests

import (
	"testing"
)

func Test_restart(t *testing.T) {
	service := createService()

	err := service.Stop()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_restart_non_existing(t *testing.T) {
	service := createRandomService()

	err := service.Stop()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = service.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
