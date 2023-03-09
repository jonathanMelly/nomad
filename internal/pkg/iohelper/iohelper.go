package iohelper

import (
	"github.com/gologme/log"
	"os"
)

// FileOrDirExists returns true if a iohelper object exists
func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		return false
	}

	return fileInfo.IsDir()
}
