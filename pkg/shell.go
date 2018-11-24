package pkg

import (
	"fmt"
	"os/exec"
)

// Installed returns a nicely-formatted error message if the given command-line tool is not installed.
func Installed(command string) error {
	_, err := exec.LookPath(command)
	if err != nil {
		err = fmt.Errorf("unable to find `%s` on your PATH. Please ensure it is installed", command)
	}
	return err
}
