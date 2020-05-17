package spec

// Supported os types
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
type OsType string

const (
	OsFreeBSD = OsType("freebsd")
	OsLinux   = OsType("linux")
	OsMacOS   = OsType("macos")
	OsWindows = OsType("windows")
	OsOpenBSD = OsType("openbsd")
	OsNetBSD  = OsType("netbsd")
)

// Supported init types
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
type InitType string

const (
	InitRC_D    = InitType("rc.d")
	InitSystemd = InitType("systemd")
	InitSystemV = InitType("systemv")
	InitUpstart = InitType("upstart")
	InitUknown  = InitType("uknown")
)

// Supported common denominator service types
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
type ServiceType string

const (
	ServiceNetwork   = ServiceType("network")
	ServiceBluetooth = ServiceType("bluetooth")
)
