package service

import (
	"errors"
)

// ErrServiceDoesNotExist -
var ErrServiceDoesNotExist = errors.New("Service does not exist")

// ErrServiceConfigError -
var ErrServiceConfigError = errors.New("Service config error")

// ErrServiceUnsupportedRequest -
var ErrServiceUnsupportedRequest = errors.New("Service unsupported request")
