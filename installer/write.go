package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// writeScripts creates files from the app-definitions file
func writeScripts(scripts map[string]string, workingFolder string) error {
	// Loop through each script
	for name, body := range scripts {

		// Path of file
		relativePath := filepath.Join(workingFolder, name)

		// Write to file
		err := ioutil.WriteFile(relativePath, []byte(body), os.ModePerm)
		if err != nil {
			return err
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
