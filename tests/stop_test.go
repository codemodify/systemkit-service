package tests

import (
	"testing"

	helpersErrors "github.com/codemodify/systemkit-helpers-reflection"
	service "github.com/codemodify/systemkit-service"
)

func Test_stop(t *testing.T) {
	systemService := CreateService()

	err := systemService.Stop()
	if helpersErrors.Is(err, service.ErrServiceDoesNotExist) {
		// this is a good thing
	} else if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_stop_non_existing(t *testing.T) {
	systemService := CreateRandomService()

	err := systemService.Stop()
	if helpersErrors.Is(err, service.ErrServiceDoesNotExist) {
		// this is a good thing
	} else if err != nil {
		t.Fatalf(err.Error())
	}
}
