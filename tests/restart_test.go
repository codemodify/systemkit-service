package tests

import (
	"testing"
)

func Test_restart(t *testing.T) {
	service := createService()

	err := service.Restart()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_restart_non_existing(t *testing.T) {
	service := createRandomService()

	err := service.Restart()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
