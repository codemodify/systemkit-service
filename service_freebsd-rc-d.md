# References
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/index.html
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-furthur.html

# Concept
```sh
#
# /etc/rc.d - originated from Luke Mewburn and the NetBSD community
#
# Idea behind BSD rc.d:
# - each "service" gets its own shell script able to start/stop/reload/status
# - the syntax is: /etc/rc.d/my-service start/stop/reload/status
# - /etc/rc          - drives the startup, it calls smaller scripts with `start` argument
# - /etc/rc.shutdown - calls smaller scripts with `stop` argument
# - /etc/rc.subr     - common operations implemented as shell functions
# - rcorder          - helps /etc/rc and /etc/rc.shutdown run the small scripts and respect dependencies
#
```

# Example 1
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-dummy.html
```sh
#!/bin/sh

. /etc/rc.subr

name="dummy"
start_cmd="${name}_start"
stop_cmd=":"

dummy_start()
{
	echo "Nothing started."
}

load_rc_config $name
run_rc_command "$1"
```

# Example 2
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-daemon.html

1. The `command` variable is meaningful to rc.subr

2. If it is set, rc.subr(8) will act according to the scenario of serving a conventional daemon
and provide arguments `start` / `stop` / `restart` / `poll` / `status`

3. The daemon will be started by running `$command` with command-line flags specified by `$mumbled_flags`.

4. `stop` must know the PID of the process to terminate it, rc.subr will scan through the list of
all processes, looking for a process with its name equal to `$procname` a variable meaningful to rc.subr(8)
and its value defaults to that of `$command`.

```sh
#!/bin/sh

. /etc/rc.subr

name=mumbled
rcvar=mumbled_enable

command="/usr/sbin/${name}"

load_rc_config $name
run_rc_command "$1"
```

# Example - Advanced
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-daemon-adv.html
```sh
#!/bin/sh

. /etc/rc.subr

name=mumbled
rcvar=mumbled_enable

command="/usr/sbin/${name}"
command_args="mock arguments > /dev/null 2>&1"						# (1)

pidfile="/var/run/${name}.pid"										# (2)

required_files="/etc/${name}.conf /usr/share/misc/${name}.rules"	# (3)

sig_reload="USR1"													# (4)

start_precmd="${name}_prestart"										# (5)
stop_postcmd="echo Bye-bye"											# (6)

extra_commands="reload plugh xyzzy"									# (7)

plugh_cmd="mumbled_plugh"											# (8)
xyzzy_cmd="echo 'Nothing happens.'"

mumbled_prestart()
{
	if checkyesno mumbled_smart; then								# (9)
		rc_flags="-o smart ${rc_flags}"								# (10)
	fi
	case "$mumbled_mode" in
	foo)
		rc_flags="-frotz ${rc_flags}"
		;;
	bar)
		rc_flags="-baz ${rc_flags}"
		;;
	*)
		warn "Invalid value for mumbled_mode"						# (11)
		return 1													# (12)
		;;
	esac
	run_rc_command xyzzy											# (13)
	return 0
}

mumbled_plugh()														# (14)
{
	echo 'A hollow voice says "plugh".'
}

load_rc_config $name
run_rc_command "$1"
```

- (1)    - arguments to `command` can be passed in `command_args`, added to the command line after `mumbled_flags`
- (2)    - create a `pidfile` so that its process can be found more easily and reliably
- (3)    - list them in `required_files` and rc.subr will check that those files do exist
- (4)    - customize signals to send to the daemon in case they differ from the well-known ones
- (5,6)  - performing additional tasks before or after the default methods is easy
- (7)    - if we would like to implement custom arguments, list them in extra_commands and provide methods to handle them
- (8,14) - our script supports two non-standard commands, `plugh` and `xyzzy` listed in `extra_commands`, now provide methods
- (9)    - a handy function named `checkyesno` is provided by rc.subr
- (10)   - we can affect the flags to be passed to `$command` by modifying `rc_flags` in `$start_precmd`
- (11)   - emit an important message that go to syslog as well, can be done with `debug` `info` `warn` `err` from `rc.subr`
- (12)   - if `argument_precmd` returns non-zero then main method will not be executed, `argument_postcmd` invoked if main returns 0
- (13)   - a script can invoke its own standard or non-standard commands if needed

# Example - Integrate script into rc.d framework
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-hookup.html

- To integrate into `rc.d`
	- install to `/etc/rc.d/SERVICE`
	- install to `/usr/local/etc/rc.d/SERVICE` for ports
	- <bsd.prog.mk> and <bsd.port.mk>
		- provide hooks for that, do not have to worry about the proper ownership and mode
	- system scripts should be installed from `src/etc/rc.d` through the Makefile found there

```sh
#!/bin/sh

# PROVIDE: mumbled oldmumble 									# (1)
# REQUIRE: DAEMON cleanvar frotz								# (2)
# BEFORE:  LOGIN												# (3)
# KEYWORD: nojail shutdown										# (4)

. /etc/rc.subr

name=mumbled
rcvar=mumbled_enable

command="/usr/sbin/${name}"
start_precmd="${name}_prestart"

mumbled_prestart()
{
	if ! checkyesno frotz_enable && \
	    ! /etc/rc.d/frotz forcestatus 1>/dev/null 2>&1; then
		force_depend frotz || return 1							# (5)
	fi
	return 0
}

load_rc_config $name
run_rc_command "$1"
```

- (1)   - declares the names of "conditions" our script provides

- (2,3) - tells `rcorder` to start DAEMON and cleanvar, then this one then LOGIN

- (4)   - `rcorder` keywords can be used to keep(-k) or sleep(-s) out some scripts.
			From all the files to be dependency sorted, `rcorder` will pick only those
			having a keyword from the keep list and not having a keyword from the skip list.
	- `rcorder` is used by `/etc/rc` and `/etc/rc.shutdown` these two scripts define the
		standard list of FreeBSD `rc.d` keywords
		- `nojail` - service is not for `jail` environment, the automatic startup and s
			hutdown procedures will ignore the script if inside a jail
		- `nostart` - service is to be started manually only
		- `shutdown` - service needs to be stopped before system shutdown
			- at shutdown `/etc/rc.shutdown` runs and assumes that most `rc.d` scripts have
			nothing to do at that time, it selectively in reverse order invokes `rc.d` scripts
			with the `shutdown` keyword only. For even faster shutdown `/etc/rc.shutdown` passes
			the `faststop` command to the scripts it runs so that they skip preliminary checks
			like the `pidfile`.

- (5)   - `force_depend` should be used with much care, better to revise the hierarchy of
			configuration variables for your `rc.d` scripts if they are interdependent.
			Our `mumbled` daemon requires that another `frotz` be started in advance.

# Example - Integrate script into rc.d framework - More flexibility
- https://www.freebsd.org/doc/en_US.ISO8859-1/articles/rc-scripting/rcng-args.html
```sh
#!/bin/sh

. /etc/rc.subr

name="dummy"
start_cmd="${name}_start"
stop_cmd=":"
kiss_cmd="${name}_kiss"
extra_commands="kiss"

dummy_start()
{
        if [ $# -gt 0 ]; then						# 1
                echo "Greeting message: $*"
        else
                echo "Nothing started."
        fi
}

dummy_kiss()
{
        echo -n "A ghost gives you a kiss"
        if [ $# -gt 0 ]; then						# 2
                echo -n " and whispers: $*"
        fi
        case "$*" in
        *[.!?])
                echo
                ;;
        *)
                echo .
                ;;
        esac
}

load_rc_config $name
run_rc_command "$@"									# 3
```
- (1) - arguments you type after `start` can end up as positional parameters to the respective method
	- ex: `/etc/rc.d/dummy start Hello world!`

- (2) - same applies to any method our script provides
	- ex: `/etc/rc.d/dummy kiss Once I was Etaoin Shrdlu...`

- (3) - if we want to pass all extra arguments to any method, substitute "$@" for "$1" in the last line
	where we invoke `run_rc_command`

