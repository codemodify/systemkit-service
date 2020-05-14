// +build !windows
// +build !darwin

package service

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
)

type initFlavor uint8

const (
	initSystemV = initFlavor(iota)
	initSystemd
	initUpstart
	initUknown
)

var logTag = "UNIX-SERVICE"

// NewServiceFromConfig -
func newServiceFromConfig(config Config) Service {
	switch getInitFlavor() {
	case initSystemV:
		return newServiceFromConfig_SystemV(config)
	case initSystemd:
		return newServiceFromConfig_SystemD(config)
	case initUpstart:
		return newServiceFromConfig_Upstart(config)
	default:
	}

	return nil
}

// NewServiceFromName -
func newServiceFromName(name string) (Service, error) {
	switch getInitFlavor() {
	case initSystemV:
		return newServiceFromName_SystemV(name)
	case initSystemd:
		return newServiceFromName_SystemD(name)
	case initUpstart:
		return newServiceFromName_Upstart(name)
	default:
	}

	return nil, nil
}

// NewServiceFromTemplate -
func newServiceFromTemplate(name string, template string) (Service, error) {
	switch getInitFlavor() {
	case initSystemV:
		return newServiceFromTemplate_SystemV(name, template)
	case initSystemd:
		return newServiceFromTemplate_SystemD(name, template)
	case initUpstart:
		return newServiceFromTemplate_Upstart(name, template)
	default:
	}

	return nil, nil
}

func getInitFlavor() initFlavor {
	initBinary, err := ioutil.ReadFile("/proc/1/cmdline")
	if err != nil {
		logging.Errorf("%s: can't find underlying system service framework, error: %s, from %s", logTag, err.Error(), helpersReflect.GetThisFuncName())
		return initUknown
	}

	// trim any nul bytes, this is present with some kernels
	init := string(bytes.TrimRight(initBinary, "\x00"))
	if strings.Contains(init, "init [") {
		return initSystemV
	}
	if strings.Contains(init, "systemd") {
		return initSystemd
	}
	if strings.Contains(init, "init") {
		// not so fast! you may think this is upstart, but it may be
		// a symlink to systemd... yeah, debian does that... ( x )
		var target string
		if len(init) > 9 && init[0:10] == "/sbin/init" {
			target, err = filepath.EvalSymlinks("/sbin/init")
		} else {
			target, err = filepath.EvalSymlinks(init)
		}
		if err == nil && strings.Contains(target, "systemd") {
			return initSystemd
		}

		return initUpstart
	}
	// failed to detect init system, falling back to sysvinit
	return initSystemV
}
