package installer

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func extractArchive(archivePath string, definition data.AppDefinition, appTargetPath string) error {

	var archiveFileSystem fs.FS
	switch definition.DownloadExtension {
	case ".zip":
		log.Debugln("ZIP archive")
		zipReader, err := zip.OpenReader(archivePath)
		if err != nil {
			return err
		}
		defer func(zipReader *zip.ReadCloser) {
			err := zipReader.Close()
			if err != nil {
				log.Warnln("Cannot close zipReader", err)
			}
		}(zipReader)
		archiveFileSystem = zipReader
	default:
		return errors.New(fmt.Sprint("Unsupported extension ", definition.DownloadExtension))
	}

	archiveDeepestRootFolder, err := guessDeepestRootFolder(archiveFileSystem)
	if err != nil {
		return err
	}

	log.Debugln("Deepest root in archive:", archiveDeepestRootFolder)
	return copyFromFS(archiveFileSystem, archiveDeepestRootFolder, appTargetPath, definition.ExtractRegex)

}

// copyFromFS will copyFromFS certain files from a fs to a target based on a regular expression
func copyFromFS(sourceFileSystem fs.FS, root string, targetDirectory string, allowRegExp *regexp.Regexp) error {

	// Create folder to copyFromFS files
	if !helper.FileOrDirExists(targetDirectory) {
		log.Debugln("Creating", targetDirectory)
		err := os.MkdirAll(targetDirectory, os.ModePerm)
		if err != nil {
			return err
		}
	}

	rootWithTrailingSlash := fmt.Sprint(root, "/")

	return fs.WalkDir(sourceFileSystem, root, func(path string, entry fs.DirEntry, err error) error {
		//Skip initial entry
		if path == root {
			return nil
		}

		relativePathInArchive := path
		if root != "." {
			relativePathInArchive, _ = strings.CutPrefix(path, rootWithTrailingSlash) // removes root from path
		}

		if !allowRegExp.MatchString(relativePathInArchive) {
			log.Traceln(relativePathInArchive, "discarded because of regex", allowRegExp.String())
			if entry.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// If the object is a directory, create it
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(targetDirectory, relativePathInArchive), os.ModePerm)
		} else {
			// Path of file directory
			basePathInTarget := filepath.Join(targetDirectory, filepath.Dir(relativePathInArchive))

			//Creates file directory
			if !helper.FileOrDirExists(basePathInTarget) {
				err = os.MkdirAll(basePathInTarget, os.ModePerm)
				if err != nil {
					return err
				}
			}

			archiveFileContent, err := fs.ReadFile(sourceFileSystem, path)
			if err != nil {
				return err
			}

			// Create the file
			targetCopy, err := os.Create(filepath.Join(targetDirectory, relativePathInArchive))
			defer func(out *os.File) {
				err := out.Close()
				if err != nil {
					log.Errorln(err)
				}
			}(targetCopy)
			if err != nil {
				return err
			}

			// Write the file
			_, err = targetCopy.Write(archiveFileContent)
			if err != nil {
				return err
			}
			return targetCopy.Close()
		}

	})

}

// Some archive contain a single folder at root, which then contains content...
// We want to avoid unnecessary sub paths...
func guessDeepestRootFolder(fsys fs.FS) (string, error) {
	filesCount := 0
	candidates := map[string]int{}
	root := "."
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if path == root {
			return nil
		}

		if !d.IsDir() {
			paths := strings.Builder{}
			split := strings.Split(path, "/")
			directoryParts := split[:len(split)-1 /*discard file*/]
			for _, dir := range directoryParts { /*zip spec asks for slash for path sep...*/
				paths.WriteString(fmt.Sprint(dir, "/"))
				candidates[paths.String()]++
			}
			filesCount++
		}

		return nil
	})
	if err != nil {
		return root, err
	}
	var winners []string //to handle multiple subdirectories...
	for candidate, viewCount := range candidates {
		if viewCount == filesCount {
			winners = append(winners, candidate[:len(candidate)-1 /*remove last trailing slash*/])
		}
	}

	length := 0
	champion := root
	for _, winner := range winners {
		newLength := len(winner)
		if newLength > length {
			champion = winner
			length = newLength
		}
	}

	return champion, nil
}
