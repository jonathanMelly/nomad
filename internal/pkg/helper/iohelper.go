package helper

import (
	"github.com/gologme/log"
	"io/fs"
	"os"
)

func stat(path string) os.FileInfo {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Errorln("Cannot stat", path, err)
	}
	return fileInfo
}

// FileOrDirExists returns true if a helper object exists
func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func IsDirectory(path string) bool {
	fileInfo := stat(path)
	return fileInfo.IsDir()
}

func IsSymlink(path string) bool {
	fileInfo := stat(path)
	return fileInfo.Mode() == fs.ModeSymlink
}
