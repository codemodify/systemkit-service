package encoders

import "github.com/codemodify/systemkit-service/spec"

var knownOsTypes = []spec.OsType{
	spec.OsFreeBSD,
	spec.OsLinux,
	spec.OsMacOS,
	spec.OsWindows,
	spec.OsOpenBSD,
	spec.OsNetBSD,
}

var knownInitTypes = []spec.InitType{
	spec.InitRC_D,
	spec.InitSystemd,
	spec.InitSystemV,
	spec.InitUpstart,
	spec.InitUknown,
}

var knownServiceTypes = []spec.ServiceType{
	spec.ServiceNetwork,
	spec.ServiceBluetooth,
}

var serviceMappings = map[spec.InitType]map[spec.ServiceType]string{

	// based on /etc/rc.d/* and /usr/local/etc/rc.d/*
	spec.InitRC_D: {
		spec.ServiceNetwork:   "NETWORKING",
		spec.ServiceBluetooth: "bluetooth",
	},

	// based on /etc/systemd/system/* and /usr/lib/systemd/system/*
	spec.InitSystemd: {
		spec.ServiceNetwork:   "network.target",
		spec.ServiceBluetooth: "bluetooth.target",
	},
}
