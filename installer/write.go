package installer

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// writeScripts creates files from the app-definitions file
func writeScripts(scripts map[string]string, workingFolder string, version string) error {
	// Loop through each script
	for name, body := range scripts {

		// Path of file
		relativePath := strings.Replace(filepath.Join(workingFolder, name), "{{VERSION}}", version, -1)

		// Write to file
		if isExist(relativePath) {
			log.Println(relativePath + " already in destination, skipping")
		} else {
			err := ioutil.WriteFile(relativePath, []byte(strings.Replace(body, "{{VERSION}}", version, -1)), os.ModePerm)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// moveObjects moves or renames files
func moveObjects(files map[string]string, workingFolder string) error {
	// Loop through each file
	for dst, src := range files {

		// Path of files
		dstFile := filepath.Join(workingFolder, dst)
		srcFile := filepath.Join(workingFolder, src)

		// Rename Files
		err := os.Rename(dstFile, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// createFolders creates folders from the CreateFolders attribute
func createFolders(folders []string, dstFolder string) error {
	// Loop through each folder
	for _, folder := range folders {

		// Path of file
		newPath := filepath.Join(dstFolder, folder)

		if !isExist(newPath) {
			err := os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//restore files from previous version (mainly for config)
func restoreFiles(files []string, source string, destination string) error {

	if destination == "" || !isExist(destination) {
		log.Println("Missing destination " + destination + " for restore => restore operation cancelled")
		return nil
	}
	if source == "" || !isExist(source) {
		log.Println("Missing source " + source + " for restore => restore operation cancelled")
		return nil
	}

	// Loop through each folder
	for _, file := range files {

		sourcePath := filepath.Join(source, file)

		if isExist(sourcePath) {
			if isDirectory(sourcePath) {
				err := filepath.Walk(sourcePath, func(walkingPath string, _ os.FileInfo, _ error) error {

					//walk walks given folder as well...
					if walkingPath != sourcePath {
						//log.Println("=>Walking " + walkingPath)
						destSubPath := strings.Join(strings.Split(walkingPath, string(os.PathSeparator))[2:], string(os.PathSeparator))
						restore(walkingPath, filepath.Join(destination, destSubPath))
					}

					return nil
				})

				if err != nil {
					log.Println("Cannot walk source "+sourcePath+", skipping|", err)
				}
			} else {
				destinationPath := filepath.Join(destination, file)
				restore(sourcePath, destinationPath)
			}
		} else {
			log.Println(sourcePath + " does not exist, skipping restore")
		}

	}

	return nil
}

//Copy source to new location
func restore(sourcePath string, destinationPath string) {
	log.Println("==> restoring " + sourcePath + " -> " + destinationPath)
	stat, err := os.Stat(sourcePath)
	if err != nil {
		log.Println("Cannot stat source "+sourcePath+", skipping|", err)
		return
	}
	bytesRead, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		log.Println("Cannot read source "+sourcePath+", skipping|", err)
	} else {
		if isExist(destinationPath) {
			log.Println("===> " + destinationPath + " already exists, trying to backup")
			newpath := destinationPath + "-" + time.Now().Format("2006-01-02-15h04m05s")
			err := os.Rename(destinationPath, newpath)
			if err != nil {
				log.Println("Cannot rename "+destinationPath+" to "+newpath+", skipping|", err)
				return
			} else {
				log.Println("===> Backuped " + destinationPath + " to " + newpath)
			}
		}

		//Warning, filepath.Dir is not os separator agnostic...
		destinationDirectory := filepath.Dir(filepath.ToSlash(destinationPath))
		if !isExist(destinationDirectory) {
			err := os.MkdirAll(destinationDirectory, os.ModePerm)
			if err != nil {
				log.Println("Cannot mkdir "+destinationDirectory+", aborting restore of "+destinationPath+" |", err)
				return
			} else {
				log.Println(destinationDirectory + " created")
			}
		}

		err = ioutil.WriteFile(destinationPath, bytesRead, stat.Mode())
		if err != nil {
			log.Println("Cannot write destination "+destinationPath+", restore failed |", err)
		} else {
			log.Println(sourcePath + " restored into " + destinationPath)
		}
	}
}
