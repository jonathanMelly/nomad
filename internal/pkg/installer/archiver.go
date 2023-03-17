package installer

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var msiDstFolder string
var msiSrcFolder string
var msiAllowRegExp *regexp.Regexp

func extractArchive(archivePath string, definition data.AppDefinition, rootFolder *string, workingFolder *string) error {
	// If RemoveRootFolder is set to true
	if definition.RemoveRootFolder {
		// If the root folder name is specified
		if len(definition.RootFolderName) > 0 {
			workingFolder = &definition.RootFolderName
		} else { // Else the root folder name is not specified so guess it
			// Return the name of the root folder in the ZIP

			workingFolderUpdated, err := discoverRootFolder(archivePath, definition.DownloadExtension)
			if err != nil {
				return errors.New(fmt.Sprint("Error discovering working folder ", err))
			} else {
				workingFolder = &workingFolderUpdated
			}
		}
	} else {
		rootFolder = workingFolder
	}

	var err error
	switch definition.DownloadExtension {
	case ".zip":
		err = extractZipRegex(archivePath, *workingFolder, definition.ExtractRegex)
	case ".msi":
		err = extractMsi(workingFolder, definition, archivePath)
	default:
		err = errors.New(fmt.Sprint("Unsupported extension ", definition.DownloadExtension))
	}

	return err
}

func extractMsi(folderName *string, definition data.AppDefinition, archivePath string) error {
	log.Debugln("MSI archive")
	// Make the folder
	err := os.Mkdir(*folderName, os.ModePerm)
	if err != nil {
		return errors.New(fmt.Sprint("Error making folder ", err))
	}

	// Get the full folder path
	fullFolderPath, err := filepath.Abs(*folderName)
	if err != nil {
		return errors.New(fmt.Sprint("Error getting folder full path ", err))
	}

	err = msiExec(archivePath, fullFolderPath, err)
	if err != nil {
		return err
	}

	// If RemoveRootFolder is set to true
	if definition.RemoveRootFolder {
		// If the root folder name is specified
		if len(definition.RootFolderName) > 0 {

			//Get the full path of the folder to set as the root folder
			currentPath := filepath.Join(fullFolderPath, definition.RootFolderName)

			// Check to make sure the path is valid
			if currentPath == fullFolderPath {
				return errors.New(fmt.Sprint("RootFolderName is invalid:", definition.RootFolderName))
			}

			// Copy files based on regular expressions
			if _, err := copyMsiRegex(currentPath, fullFolderPath+"_temp", definition.ExtractRegex); err != nil {
				return errors.New("Error restore from msi folder ")
			}

			// Set the working folder so the renaming will work later
			newWorkingFolder := fmt.Sprint(fullFolderPath, "_temp")
			folderName = &newWorkingFolder

			if err := os.RemoveAll(fullFolderPath); err != nil {
				return errors.New(fmt.Sprint("Error removing MSI folder: ", currentPath, " | ", err))
			}

		} else { // Else the root folder name is not specified
			return errors.New("tshe string, RemoveRootName, is required for MSIs")
		}
	} else {
		return errors.New("the boolean, RemoveRootFolder, is required for MSIs")
	}

	return nil
}

func msiExec(zip string, fullFolderPath string, err error) error {
	// Manually set the arguments since Go escaping does not work with MSI arguments
	argString := fmt.Sprintf(`/a "%v" /qb TARGETDIR="%v"`, zip, fullFolderPath)
	// Build the command
	log.Traceln("msi args: ", argString)
	cmd := exec.Command("msiexec", argString)

	if err = cmd.Run(); err != nil {
		return errors.New(fmt.Sprint("msi error, command: msiexec ", argString, "|", err))
	}
	return nil
}

// copyMsiRegex will restore certain files from a directory to another folder based on a regular expression
func copyMsiRegex(srcFolder string, dstFolder string, allowRegExp *regexp.Regexp) (bool, error) {
	// Create folder to extract files
	if !helper.FileOrDirExists(dstFolder) {
		err := os.MkdirAll(dstFolder, os.ModePerm)
		if err != nil {
			return false, err
		}
	}

	msiDstFolder = dstFolder
	msiSrcFolder = srcFolder
	msiAllowRegExp = allowRegExp

	err := filepath.Walk(srcFolder, visitMsiFile)
	if err != nil {
		return false, err
	}

	return true, nil
}

func visitMsiFile(fp string, _ os.FileInfo, err error) error {
	if err != nil {
		return nil // can't walk here, but continue walking elsewhere
	}

	// Path AFTER the source directory (not including the src dir)
	relativePath := strings.TrimLeft(fp, msiSrcFolder)

	// Destination path
	finalPath := filepath.Join(msiDstFolder, relativePath)

	// Destination path folder
	basePath := filepath.Dir(finalPath)

	// Check if the file matches the regular expression
	if !msiAllowRegExp.MatchString(strings.Replace(relativePath, "\\", "/", -1)) {
		return nil
	}

	// Create the file directory if it doesn't exist
	if !helper.FileOrDirExists(basePath) {
		err = os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Move the file
	err = os.Rename(fp, finalPath)
	if err != nil {
		return err
	}

	return nil
}

// extractZipRegex will extract certain files from a ZIP file to a folder based on a regular expression
func extractZipRegex(file string, rootFolder string, allowRegExp *regexp.Regexp) error {
	// Open a zip archive
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Errorln(err)
		}
	}(r)

	// If the rootFolder is NOT empty,
	if rootFolder != "" {
		// Create folder to extract files
		if !helper.FileOrDirExists(rootFolder) {
			err = os.MkdirAll(rootFolder, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	// Loop through all files
	for _, f := range r.File {

		if !allowRegExp.MatchString(f.Name) {
			continue
		}

		// Path of file
		relativePath := filepath.Join(rootFolder, f.Name)

		// Path of file directory
		basePath := filepath.Dir(relativePath)

		// If the object is a directory, create it
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(relativePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		// Create the file directory if it doesn't exist
		if !helper.FileOrDirExists(basePath) {
			err = os.MkdirAll(basePath, os.ModePerm)
			if err != nil {
				return err
			}
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		// Create the file
		out, err := os.Create(relativePath)
		defer func(out *os.File) {
			err := out.Close()
			if err != nil {
				log.Errorln(err)
			}
		}(out)
		if err != nil {
			return err
		}

		// Write the file
		_, err = io.Copy(out, rc)
		if err != nil {
			return err
		}
		err = rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func discoverRootFolder(archive string, extension string) (string, error) {
	switch extension {
	case ".zip":
		return guessZipRootFolder(archive)
	default:
		return "", errors.New("Unsupported extension " + extension)
	}
}

// guessZipRootFolder will extract a folder from a ZIP file
func guessZipRootFolder(file string) (string, error) {
	// Open a zip archive
	r, err := zip.OpenReader(file)
	if err != nil {
		return "", err
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Errorln(err)
		}
	}(r)

	if len(r.File) > 0 {
		pathArray := strings.Split(r.File[0].Name, "/")
		return pathArray[0], nil
	}

	return "", errors.New("working folder not found in first file path")
}
