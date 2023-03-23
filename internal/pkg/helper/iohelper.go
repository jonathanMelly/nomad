package helper

import (
	"errors"
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

// FileOrDirExists returns true if a file/directory is valid AND exists
func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || errors.Is(err, fs.ErrExist)
}

func IsDirectory(path string) bool {
	fileInfo := stat(path)
	return fileInfo.IsDir()
}

func IsSymlink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		log.Errorln("Cannot lstat", path, "|", err)
		return false
	}
	isSymlink := fs.ModeSymlink&fileInfo.Mode() == fs.ModeSymlink

	//Guess mor aggressively in case lstat would return bad info
	if !isSymlink {
		_, err := os.Readlink(path)
		return err == nil
	}
	return isSymlink
}
