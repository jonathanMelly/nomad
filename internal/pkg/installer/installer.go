package installer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/pkg/bytesize"
	"github.com/jonathanMelly/nomad/pkg/version"
	junction "github.com/nyaosorg/go-windows-junction"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//goland:noinspection GoSnakeCaseUsage
const (
	EXIT_OK = 0

	EXIT_INVALID_DEFINITION = 51

	EXIT_ABORTED_BY_USER = 52

	EXIT_INSTALL_UPDATE_ERROR = 53

	EXIT_MOVE_OBJECTS_ERROR   = 54
	EXIT_CREATE_FILES_ERROR   = 55
	EXIT_CREATE_FOLDERS_ERROR = 56
	EXIT_RENAME_ERROR         = 57
	EXIT_SYMLINK_ERROR        = 58
	EXIT_SHORTCUT_ERROR       = 59
)

// InstallOrUpdate will execute commands from an app-definitions file
func InstallOrUpdate(state data.AppState, forceExtract bool, skipDownload bool,
	customAppLocationForShortcut string, archivesSubDir string, askForConfirmation bool) (error error, errorMessage string, exitCode int) {

	//Aliases
	definition := state.Definition
	targetVersion := state.TargetVersion
	appName := definition.ApplicationName

	//Prepend app name to logs
	log.SetPrefix("|" + appName + "| ")

	//Enforce validation
	if !definition.Validated {
		err := definition.ValidateAndSetDefaults()
		if err != nil {
			return err, "invalid definition", EXIT_INVALID_DEFINITION
		}
	}

	//User confirm
	abort := !userWantsToContinue(askForConfirmation)
	if abort {
		return nil, "Action aborted by user", EXIT_ABORTED_BY_USER
	}

	//Create app path if needed
	if !helper.FileOrDirExists(configuration.AppPath) {
		log.Debugln("Creating", configuration.AppPath, "directory")
		err := os.Mkdir(configuration.AppPath, os.ModePerm)
		if err != nil {
			return err, fmt.Sprint("Cannot create ", configuration.AppPath), EXIT_OK
		}
	}

	var appNameWithVersion = fmt.Sprint(appName, "-", targetVersion)
	var folderName = path.Join(configuration.AppPath, appNameWithVersion)

	//Extract
	workingFolder, err := getAndExtractAppIfNeeded(state, forceExtract, skipDownload, folderName, archivesSubDir, appNameWithVersion, definition)
	if err != nil {
		return err, "Cannot install/update app", EXIT_INSTALL_UPDATE_ERROR
	}

	//Custom file actions
	err, message, code := handleCustomFileOperations(definition, workingFolder, targetVersion)
	if err != nil {
		return err, message, code
	}

	//Sets FINAL name
	if workingFolder != folderName {
		log.Debugln("Renaming ", workingFolder, " to ", folderName)
		if err := os.Rename(workingFolder, folderName); err != nil {
			return err, "Error renaming folder", EXIT_RENAME_ERROR
		}
	}

	symlink, err := handleSymlink(definition, folderName, appNameWithVersion, state)
	if err != nil {
		return err, "Symlink issue", EXIT_SYMLINK_ERROR
	}

	err = handleShortcut(definition, symlink, customAppLocationForShortcut, configuration.DefaultShortcutsDir)
	if err != nil {
		return err, fmt.Sprint("Cannot create shortcut dir ", configuration.DefaultShortcutsDir), EXIT_SHORTCUT_ERROR
	}

	return nil, "", 0

}

func handleShortcut(definition data.AppDefinition, symlink string, customAppLocationForShortcut string, shortcutDir string) error {
	if definition.Shortcut != "" {

		if !helper.FileOrDirExists(shortcutDir) {
			log.Debugln("Creating shortcutDir ", shortcutDir)
			err := os.Mkdir(shortcutDir, os.ModePerm)
			if err != nil {
				return err
			}
		}

		absSymlink, _ := filepath.Abs(symlink)

		pwd, _ := os.Getwd()

		targetForShortcut := path.Join(absSymlink, definition.Shortcut)
		if customAppLocationForShortcut != "" {
			targetForShortcut = path.Join(customAppLocationForShortcut, filepath.Base(pwd), symlink, definition.Shortcut)
		}

		absShortcutsDir, _ := filepath.Abs(shortcutDir)

		icon := ""
		if definition.ShortcutIcon != "" {
			icon = path.Join(path.Dir(targetForShortcut), definition.ShortcutIcon)
		}

		log.Debugln("Creating shortcut ", definition.Shortcut, " -> ", targetForShortcut)
		createShortcut(
			definition.Shortcut,
			targetForShortcut,
			"",
			filepath.Dir(targetForShortcut),
			fmt.Sprint("portable-", definition.Shortcut),
			absShortcutsDir,
			icon)
	}
	return nil
}

func handleCustomFileOperations(definition data.AppDefinition, workingFolder string, targetVersion *version.Version) (error, string, int) {
	log.Traceln("Creating folders:", definition.CreateFolders)
	if err := createFolders(definition.CreateFolders, workingFolder); err != nil {
		return err, "Error creating folders", EXIT_CREATE_FOLDERS_ERROR
	}

	log.Traceln("Creating files:", definition.CreateFiles)
	if err := writeScripts(definition.CreateFiles, workingFolder, targetVersion.String()); err != nil {
		return err, "Error writing files", EXIT_CREATE_FILES_ERROR
	}

	log.Traceln("Moving objects:", definition.MoveObjects)
	if err := moveObjects(definition.MoveObjects, workingFolder); err != nil {
		return err, "Error moving objects ", EXIT_MOVE_OBJECTS_ERROR
	}
	return nil, "", EXIT_OK
}

func handleSymlink(definition data.AppDefinition, folderName string, appNameWithVersion string, state data.AppState) (string, error) {
	//create/update symlink app-1.0.2 => app ...
	log.Traceln("Handling symlink and restores")

	//Can only restore from previous symlink (targetVersion)....
	symlink := path.Join(configuration.AppPath, definition.Symlink)
	if helper.FileOrDirExists(symlink) {
		absoluteFolderName, _ := filepath.Abs(folderName)
		evalSymlink, _ := filepath.EvalSymlinks(symlink)
		absoluteSymlink, _ := filepath.Abs(evalSymlink)
		if absoluteFolderName != absoluteSymlink {
			if len(definition.RestoreFiles) > 0 {
				//(handles customs/configurations that are overwritten upon upgrade)
				log.Debugln("Restoring ", definition.RestoreFiles)
				if err := restoreFiles(definition.RestoreFiles, symlink, folderName); err != nil {
					return symlink, errors.New(fmt.Sprint("Error restoring files |", err))
				}
			}
		} else {
			log.Debugln("No version change, skipping restore")
		}

		//Remove old
		err := os.Remove(symlink)
		if err != nil {
			return symlink, errors.New(fmt.Sprint("Cannot remove symlink ", symlink, " for update to latest version... | ", err))
		}
	} else if state.CurrentVersion == state.TargetVersion { /*no symlink and same version,... */
		log.Infoln("missing symlink will be regenerated")
	}

	//SYMLINK
	target := filepath.Join(configuration.AppPath, appNameWithVersion, "")
	log.Debugln("Linking " + symlink + " -> " + target)
	err := junction.Create(target, symlink)
	if err != nil {
		return symlink, errors.New(fmt.Sprint("Error symlink/junction to ", target, " | ", err))
	}

	return symlink, nil
}

func getAndExtractAppIfNeeded(
	state data.AppState,
	forceExtract bool,
	skipDownload bool,
	folderName string,
	archivesSubDir string,
	appNameWithVersion string,
	definition data.AppDefinition,
) (string, error) {

	needsExtraction, err := prepareExtractionIfNeeded(folderName, forceExtract)
	if err != nil {
		return "", err
	}

	// Working folder is the root folder where the files will be extracted
	workingFolder := folderName

	// Root folder is directory relative to the current directory where the files will be extracted to
	rootFolder := ""

	if needsExtraction {
		// Set the archivePath name based off the folder
		// Note: The original file download name will be changed
		archivePath, err := prepareArchiveDestination(archivesSubDir, appNameWithVersion, definition)
		if err != nil {
			return workingFolder, err
		}

		// Download Archive
		err = downloadArchive(state, skipDownload, archivePath, definition)
		if err != nil {
			return workingFolder, errors.New(fmt.Sprint("Cannot download archive | ", err))
		}

		//Extract
		log.Debugln("Extracting files from archive", archivePath)
		err = extractArchive(archivePath, definition, &rootFolder, &workingFolder)
		if err != nil {
			return workingFolder, errors.New(fmt.Sprint("Error extracting from archive | ", err))
		}
	} else {
		log.Infoln("directory already exists, use -force to regenerate from archive")
	}
	return workingFolder, nil
}

func prepareExtractionIfNeeded(folderName string, forceExtract bool) (bool, error) {
	log.Debugln("Extracting to ", folderName)
	extract := true
	// If the folder exists
	if helper.FileOrDirExists(folderName) {
		if helper.IsDirectory(folderName) {
			if forceExtract {
				log.Infoln("Removing old version:", folderName, " (as force extract asked)")
				err := os.RemoveAll(folderName)
				if err != nil {
					return false, errors.New(fmt.Sprint("Error removing working folder |", err))
				}
			} else {
				log.Traceln("Directory ", folderName, " already exists, letting original content unmodified (use -force)")
				extract = false
			}
		} else {
			log.Warnln("/!\\WARNINIG, filename (not directory) ", folderName, " already exists. Please remove it manually")
			extract = false
		}
	}
	return extract, nil
}

func prepareArchiveDestination(archivesSubDir string, appNameWithVersion string, definition data.AppDefinition) (string, error) {
	var archivePath = path.Join(configuration.AppPath, archivesSubDir, fmt.Sprint(appNameWithVersion, definition.DownloadExtension))
	if helper.FileOrDirExists(archivePath) {
		log.Debugln("Download Exists:", archivePath)
	} else {
		zipDir := filepath.Dir(archivePath)
		if !helper.FileOrDirExists(zipDir) {
			err := os.MkdirAll(zipDir, os.ModePerm)
			if err != nil {
				return archivePath, errors.New(fmt.Sprint("Cannot create ", zipDir))
			}
		}
	}
	return archivePath, nil
}

func downloadArchive(state data.AppState, skipDownload bool, archivePath string, definition data.AppDefinition) error {
	if skipDownload && helper.FileOrDirExists(archivePath) {
		log.Debugln("Skipping download as ", archivePath, " archive already exists (use -force to override)")
	} else {
		downloadURL := state.TargetVersion.FillVersionsPlaceholders(definition.DownloadUrl)
		log.Debugln("Downloading from ", downloadURL, " >> ", archivePath)

		if strings.HasPrefix(downloadURL, "manual") {
			scanner := bufio.NewScanner(os.Stdin)
			log.Print("Please paste custom URL for download (", downloadURL, ") :")
			scanner.Scan()
			answer := scanner.Text()
			log.Debugln("Custom URL", answer)
			downloadURL = answer
		}

		size, err := helper.DownloadFile(downloadURL, archivePath)
		if err != nil {
			return errors.New(fmt.Sprint("Error download file ", err))
		}
		log.Traceln("Download size:", bytesize.ByteSize(size))
	}
	return nil
}

func userWantsToContinue(askForConfirmation bool) bool {
	if askForConfirmation {
		scanner := bufio.NewScanner(os.Stdin)
		log.Print("Proceed [Y,n] ?")
		scanner.Scan()
		answer := scanner.Text()
		log.Debugln("Answer", answer)
		return answer == "" || strings.ToLower(answer) == "y"

	}
	return false
}
