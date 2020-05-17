// +build linux

package service

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
	logging "github.com/codemodify/systemkit-logging"
	"github.com/codemodify/systemkit-service/spec"
)

var logTag = "LINUX-SERVICE"

func newServiceFromSERVICE(serviceSpec spec.SERVICE) Service {
	switch getInitType() {
	case spec.InitSystemV:
		return newServiceFromSERVICE_SystemV(serviceSpec)
	case spec.InitSystemd:
		return newServiceFromSERVICE_SystemD(serviceSpec)
	case spec.InitUpstart:
		return newServiceFromSERVICE_Upstart(serviceSpec)
	default:
	}

	return nil
}

func newServiceFromName(name string) (Service, error) {
	switch getInitType() {
	case spec.InitSystemV:
		return newServiceFromName_SystemV(name)
	case spec.InitSystemd:
		return newServiceFromName_SystemD(name)
	case spec.InitUpstart:
		return newServiceFromName_Upstart(name)
	default:
	}

	return nil, nil
}

func newServiceFromPlatformTemplate(name string, template string) (Service, error) {
	switch getInitType() {
	case spec.InitSystemV:
		return newServiceFromPlatformTemplate_SystemV(name, template)
	case spec.InitSystemd:
		return newServiceFromPlatformTemplate_SystemD(name, template)
	case spec.InitUpstart:
		return newServiceFromPlatformTemplate_Upstart(name, template)
	default:
	}

	return nil, nil
}

func getInitType() spec.InitType {
	initBinary, err := ioutil.ReadFile("/proc/1/cmdline")
	if err != nil {
		logging.Errorf("%s: can't find underlying system service framework, error: %s, from %s", logTag, err.Error(), helpersReflect.GetThisFuncName())
		return spec.InitUknown
	}

	// trim any nul bytes, this is present with some kernels
	init := string(bytes.TrimRight(initBinary, "\x00"))
	if strings.Contains(init, "init [") {
		return spec.InitSystemV
	}
	if strings.Contains(init, "systemd") {
		return spec.InitSystemd
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
			return spec.InitSystemd
		}

		return spec.InitUpstart
	}
	// failed to detect init system, falling back to sysvinit
	return spec.InitSystemV
}
