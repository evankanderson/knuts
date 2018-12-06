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

// Kubectl applies a yaml file with kubectl. Kubectl is not amenable to being called as a library.
func Kubectl(file string) error {
	cmd := exec.Command("kubectl", "apply", "--filename", file)
	if DryRun {
		fmt.Printf("Dry run: `kubectl --filename %q`\n", file)
		return nil
	}
	return cmd.Run()
}

// KubectlInline applies suplied yaml contents with kubectl. Kubectl is not amenable to being called as a library.
func KubectlInline(contents []byte) error {
	cmd := exec.Command("kubectl", "apply", "--filename", "-")
	if DryRun {
		fmt.Printf("Dry run: `kubectl < INPUT`\n")
		return nil
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Input to kubectl failed: %v", err)
	}
	go func() {
		defer in.Close()
		in.Write(contents)
	}()

	return cmd.Run()
}
