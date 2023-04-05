package installer

import (
	"errors"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/pkg/version"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// writeScripts creates files from the app-definitions file
func writeScripts(scripts map[string]string, appSpecificVersionFolder string, absoluteSymlinkToAppFolder string, version *version.Version) error {
	var _errors []error
	// Loop through each script
	for name, body := range scripts {

		// Path of file
		relativePath := strings.Replace(filepath.Join(appSpecificVersionFolder, name), "{{VERSION}}", version.String(), -1)

		// Write to file
		if helper.FileOrDirExists(relativePath) {
			log.Debugln(relativePath, "already in destination, skipping")
		} else {
			content := strings.Replace(body, "{{VERSION}}", version.String(), -1)
			content = strings.Replace(content, "{{APP_PATH_GENERIC}}", absoluteSymlinkToAppFolder, -1)
			content = strings.Replace(content, "{{APP_PATH}}", appSpecificVersionFolder, -1)
			err := os.WriteFile(relativePath, []byte(content), os.ModePerm)
			if err != nil {
				_errors = append(_errors, err)
			}
		}

	}

	return errors.Join(_errors...)

}

// moveObjects moves or renames files
func moveObjects(files map[string]string, workingFolder string) error {
	var _errors []error
	// Loop through each file
	for dst, src := range files {

		// Path of files
		dstFile := filepath.Join(workingFolder, dst)
		srcFile := filepath.Join(workingFolder, src)

		// Rename Files
		err := os.Rename(dstFile, srcFile)
		if err != nil {
			_errors = append(_errors, err)
		}
	}

	return errors.Join(_errors...)
}

// createFolders creates folders from the CreateFolders attribute
func createFolders(folders []string, dstFolder string) error {
	var _errors []error
	// Loop through each folder
	for _, folder := range folders {

		// Path of file
		newPath := filepath.Join(dstFolder, folder)

		if !helper.FileOrDirExists(newPath) {
			err := os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				_errors = append(_errors, err)
			}
		}
	}

	return errors.Join(_errors...)
}

// restore files from previous version (mainly for config)
func restoreFiles(files []string, source string, destination string) error {

	if destination == "" || !helper.FileOrDirExists(destination) {
		log.Debugln("Missing destination", destination, "for restore => restore operation cancelled")
		return nil
	}
	if source == "" || !helper.FileOrDirExists(source) {
		log.Debugln("Missing source", source, " for restore => restore operation cancelled")
		return nil
	}

	// Loop through each folder
	for _, file := range files {

		sourcePath := filepath.Join(source, file)

		if helper.FileOrDirExists(sourcePath) {
			if helper.IsDirectory(sourcePath) {
				err := filepath.Walk(sourcePath, func(walkingPath string, fileInfo os.FileInfo, _ error) error {

					//walk walks given folder as well...
					if walkingPath != sourcePath && !fileInfo.IsDir() {
						log.Traceln("=>Walking " + walkingPath)
						destSubPath := strings.Join(strings.Split(walkingPath, string(os.PathSeparator))[2:], string(os.PathSeparator))
						restore(walkingPath, filepath.Join(destination, destSubPath))
					}

					return nil
				})

				if err != nil {
					log.Warnln("Cannot walk source "+sourcePath+", skipping|", err)
				}
			} else {
				destinationPath := filepath.Join(destination, file)
				restore(sourcePath, destinationPath)
			}
		} else {
			log.Warnln(sourcePath, "does not exist, skipping restore")
		}

	}

	return nil
}

// Copy source to upcoming location
func restore(sourcePath string, destinationPath string) {
	log.Debugln("==> restoring " + sourcePath + " -> " + destinationPath)
	stat, err := os.Stat(sourcePath)
	if err != nil {
		log.Errorln("Cannot stat source ", sourcePath, ", skipping|", err)
		return
	}
	bytesRead, err := os.ReadFile(sourcePath)
	if err != nil {
		log.Errorln("Cannot read source ", sourcePath, ", skipping|", err)
	} else {
		if helper.FileOrDirExists(destinationPath) {
			log.Debugln(destinationPath, " already exists, trying to backup")
			newpath := destinationPath + "-" + time.Now().Format("2006-01-02-15h04m05s")
			err := os.Rename(destinationPath, newpath)
			if err != nil {
				log.Errorln("Cannot rename ", destinationPath, "to", newpath, ", skipping|", err)
				return
			} else {
				log.Debugln("Backuped ", destinationPath, " to ", newpath)
			}
		}

		//Warning, filepath.Dir is not os separator agnostic...
		destinationDirectory := filepath.Dir(filepath.ToSlash(destinationPath))
		if !helper.FileOrDirExists(destinationDirectory) {
			err := os.MkdirAll(destinationDirectory, os.ModePerm)
			if err != nil {
				log.Errorln("Cannot mkdir ", destinationDirectory, ", aborting restore of ", destinationPath, " |", err)
				return
			} else {
				log.Debugln(destinationDirectory, " created")
			}
		}

		err = os.WriteFile(destinationPath, bytesRead, stat.Mode())
		if err != nil {
			log.Errorln("Cannot write destination ", destinationPath, ", restore failed |", err)
		} else {
			log.Debugln(sourcePath, " restored into ", destinationPath)
		}
	}
}
