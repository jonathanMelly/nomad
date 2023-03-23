package installer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/internal/pkg/state"
	"github.com/jonathanMelly/nomad/pkg/bytesize"
	junction "github.com/nyaosorg/go-windows-junction"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

//goland:noinspection GoSnakeCaseUsage
const (
	EXIT_OK = 0

	EXIT_INVALID_DEFINITION = 51

	EXIT_ABORTED_BY_USER = 52

	EXIT_INSTALL_UPDATE_ERROR = 53

	EXIT_SYMLINK_ERROR  = 58
	EXIT_SHORTCUT_ERROR = 59
)

// InstallOrUpdate will execute commands from an app-definitions file
func InstallOrUpdate(appState state.AppState, forceExtract bool, skipDownload bool,
	customAppLocationForShortcut string, archivesSubDir string, askForConfirmation bool) (error error, errorMessage string, exitCode int) {

	//Aliases
	definition := appState.Definition
	targetVersion := appState.TargetVersion
	appName := definition.ApplicationName

	//Prepend app name to logs
	log.SetPrefix(helper.BuildPrefix(appName))

	//Enforce validation
	if !definition.Validated {
		err := definition.ValidateAndSetDefaults()
		if err != nil {
			return err, "invalid definition", EXIT_INVALID_DEFINITION
		}
	}

	//Show status
	log.Infoln(appState.StatusMessage())

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
	var targetAppPath = path.Join(configuration.AppPath, appNameWithVersion)
	var archivesDir = path.Join(configuration.AppPath, archivesSubDir)

	//Extract
	err := getAndExtractAppIfNeeded(appState, forceExtract, skipDownload, targetAppPath, archivesDir, appNameWithVersion, &definition)
	if err != nil {
		return err, "Cannot install/update app", EXIT_INSTALL_UPDATE_ERROR
	}

	//Custom file actions
	handleRestoreAndCustomFiles(appState, targetAppPath)

	//Symlink
	symlink, err := handleSymlink(appState, targetAppPath)
	if err != nil {
		return err, "Symlink issue", EXIT_SYMLINK_ERROR
	}

	//Shortcut
	err = handleShortcut(definition, symlink, customAppLocationForShortcut, configuration.DefaultShortcutsDir)
	if err != nil {
		return err, fmt.Sprint("Cannot create shortcut dir ", configuration.DefaultShortcutsDir), EXIT_SHORTCUT_ERROR
	}

	log.Infoln(appState.SuccessMessage())
	return nil, "", EXIT_OK

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

func handleRestoreAndCustomFiles(appState state.AppState, workingFolder string) {

	definition := appState.Definition
	if appState.Status != state.KEEP {
		log.Traceln("Restoring files folders:", definition.RestoreFiles)
		if err := restoreFiles(definition.RestoreFiles, appState.CurrentVersionFolder, workingFolder); err != nil {
			log.Errorln("Error restoring files:", definition.RestoreFiles, "|", err)
		}
	} else {
		log.Debugln("No version change, skipping restore")
	}

	log.Traceln("Creating folders:", definition.CreateFolders)
	if err := createFolders(definition.CreateFolders, workingFolder); err != nil {
		log.Errorln("Error creating folders:", definition.CreateFolders, "|", err)
	}

	log.Traceln("Creating files:", definition.CreateFiles)
	if err := writeScripts(definition.CreateFiles, workingFolder, appState.TargetVersion.String()); err != nil {
		log.Errorln("Error creating files:", definition.CreateFiles, "|", err)
	}

	log.Traceln("Moving objects:", definition.MoveObjects)
	if err := moveObjects(definition.MoveObjects, workingFolder); err != nil {
		log.Errorln("Error moving objects:", definition.MoveObjects, "|", err)
	}
}

func handleSymlink(appState state.AppState, newTarget string) (string, error) {
	//create/update symlink app-1.0.2 => app ...
	symlink := filepath.Join(configuration.AppPath, appState.Definition.Symlink)
	log.Debugln("Handling symlink", symlink, "(already discovered:", appState.SymlinkFound, ")")

	if appState.SymlinkFound /*&& helper.FileOrDirExists(symlink) should not be possible...*/ {
		//Remove old
		err := os.Remove(symlink)
		if err != nil {
			log.Errorln("Cannot remove symlink ", symlink, "|", err)
		} else {
			log.Debugln("Removed symlink", symlink)
		}
	} else if reflect.DeepEqual(appState.CurrentVersion, appState.TargetVersion) { /*no symlink and same version,... */
		log.Infoln("missing symlink", symlink, "will be regenerated")
	}

	//SYMLINK
	log.Debugln("Linking " + symlink + " -> " + newTarget)
	err := junction.Create(newTarget, symlink)
	if err != nil {
		return symlink, errors.New(fmt.Sprint("Error symlink/junction to ", newTarget, " | ", err))
	}

	return symlink, nil
}

func getAndExtractAppIfNeeded(
	appState state.AppState,
	forceExtract bool,
	skipDownload bool,
	targetAppPath string,
	archivesDir string,
	appNameWithVersion string,
	definition *data.AppDefinition,
) error {

	needsExtraction, err := checkAndEraseCurrentVersionIfNeeded(targetAppPath, forceExtract)
	if err != nil {
		return err
	}

	if needsExtraction {
		log.Debugln("Preparing for extraction")

		//Create archives base directory if needed (only first time)
		if !helper.FileOrDirExists(archivesDir) {
			log.Traceln("Creating archive dir", archivesDir)
			err := os.MkdirAll(archivesDir, os.ModePerm)
			if err != nil {
				return errors.New(fmt.Sprint("Cannot create ", archivesDir))
			}
		}

		//Get downloadURL (from human if needed)
		downloadURL := appState.TargetVersion.FillVersionsPlaceholders(definition.DownloadUrl)
		if strings.HasPrefix(downloadURL, "manual") {
			scanner := bufio.NewScanner(os.Stdin)
			log.Print("Please paste custom URL for download (", downloadURL, ") :")
			scanner.Scan()
			answer := scanner.Text()
			log.Debugln("Custom URL", answer)

			downloadURL = appState.TargetVersion.FillVersionsPlaceholders(answer)
			definition.DownloadUrl = downloadURL
			definition.ComputeDownloadExtension()
		}

		// Set the archivePath name based off the folder
		// Note: The original file download name will be changed
		var archivePath = path.Join(archivesDir, fmt.Sprint(appNameWithVersion, definition.DownloadExtension))
		log.Infoln("Downloading ", downloadURL, " >> ", archivePath)
		err = downloadArchive(downloadURL, skipDownload, archivePath)
		if err != nil {
			return errors.New(fmt.Sprint("Cannot download archive | ", err))
		}

		//Extract
		log.Debugln("Extracting files from ", archivePath)
		err = extractArchive(archivePath, *definition, targetAppPath)
		if err != nil {
			return errors.New(fmt.Sprint("Error extracting from archive | ", err))
		}
	} else {
		log.Infoln("directory already exists, use -force to regenerate from archive")
	}
	return nil
}

func checkAndEraseCurrentVersionIfNeeded(appPath string, forceExtract bool) (bool, error) {
	log.Debugln("Checking ", appPath)
	extract := true
	// If the folder exists
	if helper.FileOrDirExists(appPath) {
		if helper.IsDirectory(appPath) {
			if forceExtract {
				log.Infoln("Removing old version:", appPath, " (as force extract asked)")
				err := os.RemoveAll(appPath)
				if err != nil {
					return false, errors.New(fmt.Sprint("Error removing working folder |", err))
				}
			} else {
				log.Traceln("Directory ", appPath, " already exists, letting original content unmodified (use -force)")
				extract = false
			}
		} else {
			log.Warnln("/!\\WARNINIG, filename (not directory) ", appPath, " already exists. Please remove it manually")
			extract = false
		}
	}
	return extract, nil
}

func downloadArchive(downloadURL string, skipDownload bool, archivePath string) error {
	if skipDownload && helper.FileOrDirExists(archivePath) {
		log.Debugln("Skipping download as ", archivePath, " archive already exists (use -force to override)")
	} else {
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
		log.Print("Proceed [Y,n] (Enter=Yes) ? ")
		scanner.Scan()
		answer := scanner.Text()
		log.Debugln("Answer", answer)
		return answer == "" || strings.ToLower(answer) == "y"

	}
	return false
}
