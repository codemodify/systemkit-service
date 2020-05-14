# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) Service
[![](https://img.shields.io/github/v/release/codemodify/systemkit-service?style=flat-square)](https://github.com/codemodify/systemkit-service/releases/latest)
![](https://img.shields.io/github/languages/code-size/codemodify/systemkit-service?style=flat-square)
![](https://img.shields.io/github/last-commit/codemodify/systemkit-service?style=flat-square)
[![](https://img.shields.io/badge/license-0--license-brightgreen?style=flat-square)](https://github.com/codemodify/TheFreeLicense)

![](https://img.shields.io/github/workflow/status/codemodify/systemkit-service/qa?style=flat-square)
![](https://img.shields.io/github/issues/codemodify/systemkit-service?style=flat-square)
[![](https://goreportcard.com/badge/github.com/codemodify/systemkit-service?style=flat-square)](https://goreportcard.com/report/github.com/codemodify/systemkit-service)

[![](https://img.shields.io/badge/godoc-reference-brightgreen?style=flat-square)](https://godoc.org/github.com/codemodify/systemkit-service)
![](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)
![](https://img.shields.io/gitter/room/codemodify/systemkit-service?style=flat-square)

![](https://img.shields.io/github/contributors/codemodify/systemkit-service?style=flat-square)
![](https://img.shields.io/github/stars/codemodify/systemkit-service?style=flat-square)
![](https://img.shields.io/github/watchers/codemodify/systemkit-service?style=flat-square)
![](https://img.shields.io/github/forks/codemodify/systemkit-service?style=flat-square)


#### Robust Cross platform Create/Start/Stop/Delete system or user service

# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) Install
```go
go get github.com/codemodify/systemkit-service
```

&nbsp;							| &nbsp; 																	| &nbsp;
---:							| ---																		| ---
__systemd__						| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://img.icons8.com/color/48/000000/raspberry-pi.png" /> <img src="https://upload.wikimedia.org/wikipedia/commons/a/a5/Archlinux-icon-crystal-64.svg" width="48" /> <img src="https://img.icons8.com/color/48/000000/debian.png"/> <img src="https://img.icons8.com/color/48/000000/ubuntu--v1.png"/> <img src="https://img.icons8.com/color/48/000000/suse.png"/> <img src="https://img.icons8.com/color/48/000000/centos.png"/> <img src="https://upload.wikimedia.org/wikipedia/commons/3/3f/Fedora_logo.svg" width="40" /> <img src="https://img.icons8.com/color/48/000000/red-hat.png"/> <img src="https://img.icons8.com/color/48/000000/linux-mint.png"/> <img src="https://img.icons8.com/color/48/000000/mandriva.png"/>
__rc.d__						| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://upload.wikimedia.org/wikipedia/en/thumb/d/df/Freebsd_logo.svg/2880px-Freebsd_logo.svg.png" width="100" /> <img src="https://www.netbsd.org/images/NetBSD-tb.png" width="50" /> <img src="https://upload.wikimedia.org/wikipedia/en/8/83/OpenBSD_Logo_-_Cartoon_Puffy_with_textual_logo_below.svg" width="80" />
__procd__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/commons/9/92/Openwrt_Logo.svg" width="150" /> <img src="https://pulpstone.pw/wp-content/uploads/lede_574-423-e1510414969868.png" width="100" />
__sysvinit__					| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://img.icons8.com/color/48/000000/linux.png" />
__launchd__						| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://img.icons8.com/color/48/000000/mac-os.png"/>
__Service Control Manager__		| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://img.icons8.com/color/48/000000/windows-10.png"/>
__cygserver__					| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Cygwin_logo.svg" width="40" />
__OpenRC__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/commons/4/48/Gentoo_Linux_logo_matte.svg" width="40" />
__Shepherd__					| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/commons/f/f6/Hurd-logo.svg" width="40" />
__Mudur__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	|
__init__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	|
__cinit__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	|
__runit__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/commons/0/02/Void_Linux_logo.svg" width="48" />
__minit__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	|
__Initng__						| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| Berry Linux
__Android Init__				| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://img.icons8.com/color/48/000000/android-os.png"/>
__UpStart__						| <img src="https://img.icons8.com/color/30/000000/verified-account.png" />	| <img src="https://img.icons8.com/color/48/000000/chrome--v1.png"/> <img src="https://img.icons8.com/color/48/000000/ubuntu--v1.png"/>
__Service Management Facility__	| <img src="https://img.icons8.com/color/30/000000/in-progress--v1.png"  />	| <img src="https://upload.wikimedia.org/wikipedia/en/8/89/IllumosPhoenixRGB.png" width="40" /> <img src="https://upload.wikimedia.org/wikipedia/commons/e/ee/Aktualne_logo_Oracle_Solaris_OS_OSos.png" width="110" />


# ![](https://fonts.gstatic.com/s/i/materialicons/bookmarks/v4/24px.svg) References
- https://en.wikipedia.org/wiki/Operating_system_service_management
- https://nosystemd.org
- https://ungleich.ch/en-us/cms/blog/2019/05/20/linux-distros-without-systemd
- https://lwn.net/Articles/578209/
- https://lwn.net/Articles/578210/
- https://en.wikipedia.org/wiki/Init
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/linux-users/startup.html
- https://sosheskaz.github.io/tutorial/2017/03/28/FreeBSD-rcd-Setup.html
