package helpers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"os/user"
)

func AsJSONString(i interface{}) string {
	bytes, err := json.Marshal(i)
	if err != nil {
		return fmt.Sprintf("ERROR: AsJSONString(), details [%s]", err.Error())
	}
	return string(bytes)
}

func Is(err1 error, err2 error) bool {
	if err1 == err2 {
		return true
	}

	if err1 != nil && err2 != nil {
		return err1.Error() == err2.Error()
	}

	return false
}

func ExecWithArgs(name string, args ...string) (out string, err error) {
	output, err := exec.Command(name, args...).CombinedOutput()
	return string(output), err
}


func IsRoot() bool {
	u, err := user.Current()

	if err != nil {
		return false
	}

	// On unix systems, root user either has the UID 0,
	// the GID 0 or both.
	return u.Uid == "0" || u.Gid == "0"
}

func HomeDir(returnIfError string) string {
	u, err := user.Current()
	if err != nil {
		return returnIfError
	}

	return u.HomeDir
}
