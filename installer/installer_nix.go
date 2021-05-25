//+build linux darwin !windows

package installer

import (
	"errors"
	"os"
)

//Install funcation installs the go binary into the host
func Install() (err error) {
	return errors.New("Aw Snap, you need to install it")
}

//SetPurge delete file from disk on exit
func SetPurge(file string) error {
	defer os.Remove(file)
	return nil
}
