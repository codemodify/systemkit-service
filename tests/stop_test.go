package tests

import (
	"testing"

	helpersErrors "github.com/codemodify/systemkit-helpers"
	service "github.com/codemodify/systemkit-service"
)

func Test_stop(t *testing.T) {
	systemService := CreateRemoteitService()

	err := systemService.Stop()
	if helpersErrors.Is(err, service.ErrServiceDoesNotExist) {
		// this is a good thing
	} else if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_stop_non_existing(t *testing.T) {
	systemService := createRandomService()

	err := systemService.Stop()
	if helpersErrors.Is(err, service.ErrServiceDoesNotExist) {
		// this is a good thing
	} else if err != nil {
		t.Fatalf(err.Error())
	}
}
