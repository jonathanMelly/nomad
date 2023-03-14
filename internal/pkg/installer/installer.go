package installer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/iohelper"
	"github.com/jonathanMelly/nomad/internal/pkg/state"
	"github.com/jonathanMelly/nomad/pkg/bytesize"
	"github.com/jonathanMelly/nomad/pkg/version"
	junction "github.com/nyaosorg/go-windows-junction"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var AppPath = "apps"

func init() {
	// Verbose logging with file name and line number
	//log.SetFlags(log.Lshortfile)
}

// Run will execute commands from an app-definitions file
func Run(action string, app data.AppDefinition, versionOverwrite string, forceExtract bool, skipDownload bool,
	customAppLocationForShortcut string, archivesSubDir string, useLatestVersion bool, confirm bool) (error error, errorMessage string, exitCode int) {

	err := app.Validate()
	if err != nil {
		log.Errorln("Invalid app definition", err)
		return
	}

	// Compile the regular expressions into one
	re, err := combineRegex(app.ExtractRegExList)
	if err != nil {
		return err, "Error creating regular expression from list", 2
	}

	appName := app.ApplicationName
	log.SetPrefix("|" + appName + "| ")

	//VERSION HANDLING
	var targetVersion *version.Version = nil

	// Overwrite the targetVersion if available
	var versionOverwriteVersion *version.Version = nil
	if versionOverwrite != "" {
		versionOverwriteVersion, err = version.FromString(versionOverwrite)
		if err != nil {
			return err, "Bad forced version format: " + app.Version, 31
		}
		log.Debugln("Version flag overwrite:", versionOverwriteVersion)
	}

	configVersion, err := version.FromString(app.Version)
	if err != nil {
		return err, "Bad version format in config :" + app.Version, 30
	}
	log.Debugln("Version from config: ", configVersion)

	//Current version
	var currentInstalledVersion *version.Version = nil
	var symlink string
	if app.Symlink != "" {
		log.Debugln("Using custom symlink", app.Symlink)
		symlink = path.Join(AppPath, app.Symlink)
	} else {
		symlink = appName
	}

	if iohelper.FileOrDirExists(symlink) {
		target, _ := os.Readlink(symlink)
		split := strings.Split(filepath.Base(target), "-")
		if len(split) > 1 {
			currentInstalledVersion, err = version.FromString(split[1])
		}
	} else {
		apps := state.GetCurrentApps(".")
		cappVersion, exist := apps[appName]
		if exist {
			currentInstalledVersion = cappVersion
		}
	}
	log.Debugln("Version installed: ", currentInstalledVersion)

	var latestVersionFromRemote *version.Version = nil
	// If Version Check parameters are specified
	if useLatestVersion && app.VersionCheck.Url != "" && app.VersionCheck.RegEx != "" {

		// Extract the targetVersion from the webpage
		latestVersionFromRemote, err = extractFromRequest(app.VersionCheck.Url, app.VersionCheck.RegEx)
		if err != nil {
			return err, "Error retrieving last version from remote", 3
		}
		log.Debugln("Version from remote: ", latestVersionFromRemote)

	}

	if versionOverwriteVersion != nil {
		targetVersion = versionOverwriteVersion
	} else {
		//not yet installed
		if currentInstalledVersion == nil {
			if latestVersionFromRemote != nil && latestVersionFromRemote.IsNewerThan(*configVersion) {
				targetVersion = latestVersionFromRemote
			} else {
				targetVersion = configVersion
			}
		} else /*Already installed*/ {
			if latestVersionFromRemote != nil && latestVersionFromRemote.IsNewerThan(*configVersion) && latestVersionFromRemote.IsNewerThan(*currentInstalledVersion) {
				targetVersion = latestVersionFromRemote
			} else if configVersion.IsNewerThan(*currentInstalledVersion) {
				targetVersion = configVersion
			} else {
				targetVersion = currentInstalledVersion
			}
		}
	}

	currentInstalledVersionString := "not installed"
	if currentInstalledVersion != nil {
		currentInstalledVersionString = currentInstalledVersion.String()
	}
	if currentInstalledVersionString != targetVersion.String() {

		if currentInstalledVersion == nil {
			action := "not installed >> will install"
			log.Infoln(action, "version", targetVersion)
		} else {
			action := "upgrading"
			if currentInstalledVersion.IsNewerThan(*targetVersion) {
				action = "downgrading"
			}
			log.Infoln(action, "version from", currentInstalledVersionString, ">>", targetVersion)
		}

	} else {
		log.Infoln("staying at version", targetVersion)
	}

	switch action {
	case "install":
		log.Debugln("Target version: ", targetVersion)
		if confirm {
			scanner := bufio.NewScanner(os.Stdin)
			log.Print("Proceed [Y,n] ?")
			scanner.Scan()
			answer := scanner.Text()
			if !(answer == "" || strings.ToLower(answer) == "y") {
				return nil, "Action aborted by user", 10
			}
		}

		if !iohelper.FileOrDirExists(AppPath) {
			log.Debugln("Creating", AppPath, "directory")
			err := os.Mkdir(AppPath, os.ModePerm)
			if err != nil {
				return err, fmt.Sprint("Cannot create ", AppPath), 0
			}
		}

		var appNameWithVersion = fmt.Sprint(appName, "-", targetVersion)

		var folderName = path.Join(AppPath, appNameWithVersion)

		// Set the zip name based off the folder
		// Note: The original file download name will be changed
		var zip = path.Join(AppPath, archivesSubDir, fmt.Sprint(appNameWithVersion, app.DownloadExtension))

		// If the zip file DOES exist on disk
		if iohelper.FileOrDirExists(zip) {
			// Output the filename of the folder
			log.Debugln("Download Exists:", zip)
		} else {
			zipDir := filepath.Dir(zip)
			if !iohelper.FileOrDirExists(zipDir) {
				err := os.MkdirAll(zipDir, os.ModePerm)
				if err != nil {
					return err, fmt.Sprint("Cannot create ", zipDir), 0
				}
			}
		}

		// If SkipDownload is true
		if skipDownload && iohelper.FileOrDirExists(zip) {
			log.Debugln("Skipping download as ", zip, " archive already exists (use -force to override)")
		} else {
			downloadURL := targetVersion.FillVersionsPlaceholders(app.DownloadUrl)

			log.Debugln("Downloading from ", downloadURL, " >> ", zip)

			size, err := downloadFile(downloadURL, zip)
			if err != nil {
				return err, "Error download file", 4
			}
			log.Traceln("Download size:", bytesize.ByteSize(size))
		}

		log.Debugln("Extracting to ", folderName)
		extract := true
		// If the folder exists
		if iohelper.FileOrDirExists(folderName) {
			if iohelper.IsDirectory(folderName) {
				if forceExtract {
					log.Debugln("Removing old folder:", folderName, " (as force extract asked)")
					err = os.RemoveAll(folderName)
					if err != nil {
						return err, "Error removing working folder |", 5
					}

				} else {
					log.Traceln("Directory ", folderName, " already exists, letting original content unmodified (use -f to force)")
					extract = false
				}
			} else {
				log.Warnln("/!\\WARNINIG, filename (not directory) ", folderName, " already exists. Please remove it manually")
				extract = false
			}

		}

		// Working folder is the root folder where the files will be extracted
		workingFolder := folderName

		// Root folder is directory relative to the current directory where the files
		// will be extracted to
		rootFolder := ""

		if extract {
			log.Debugln("Extracting files")

			switch app.DownloadExtension {
			case ".zip", ".7zTODO":
				log.Debugln("ZIP archive")
				// If RemoveRootFolder is set to true
				if app.RemoveRootFolder {
					// If the root folder name is specified
					if len(app.RootFolderName) > 0 {
						workingFolder = app.RootFolderName
					} else { // Else the root folder name is not specified so guess it
						// Return the name of the root folder in the ZIP

						workingFolder, err = extractRootFolder(zip, app.DownloadExtension)
						if err != nil {
							return err, "Error discovering working folder ", 6
						}
					}
				} else {
					rootFolder = workingFolder
				}

				// Extract files based on regular expression
				_, err = extractRegex(app.DownloadExtension, zip, rootFolder, re)
				if err != nil {
					return err, "Error extracting from zip |", 7
				}
			case ".msi":
				log.Debugln("MSI archive")
				// Make the folder
				err = os.Mkdir(folderName, os.ModePerm)
				if err != nil {
					return err, "Error making folder", 8
				}

				// Get the full folder path
				fullFolderPath, err := filepath.Abs(folderName)
				if err != nil {
					return err, "Error getting folder full path ", 9
				}

				err2, s, i, done := msiExec(zip, fullFolderPath, err)
				if done {
					return err2, s, i
				}

				// If RemoveRootFolder is set to true
				if app.RemoveRootFolder {
					// If the root folder name is specified
					if len(app.RootFolderName) > 0 {

						//Get the full path of the folder to set as the root folder
						currentPath := filepath.Join(fullFolderPath, app.RootFolderName)

						// Check to make sure the path is valid
						if currentPath == fullFolderPath {
							return nil, "RootFolderName is invalid:" + app.RootFolderName, 12
						}

						// Copy files based on regular expressions
						if _, err := copyMsiRegex(currentPath, fullFolderPath+"_temp", re); err != nil {
							return err, "Error restore from msi folder |", 13
						}

						// Set the working folder so the renaming will work later
						workingFolder = fullFolderPath + "_temp"

						if err := os.RemoveAll(fullFolderPath); err != nil {
							return err, "Error removing MSI folder:" + currentPath, 14
						}

					} else { // Else the root folder name is not specified
						return nil, "The string, RemoveRootName, is required for MSIs", 15
					}
				} else {
					return nil, "The boolean, RemoveRootFolder, is required for MSIs", 16
				}
			default:
				log.Println()
				return nil, "Download extension not supported:" + app.DownloadExtension, 17
			}
		}

		log.Traceln("Creating folders:", app.CreateFolders)
		if err := createFolders(app.CreateFolders, workingFolder); err != nil {
			return err, "Error creating folders", 18
		}

		log.Traceln("Creating files:", app.CreateFiles)
		if err := writeScripts(app.CreateFiles, workingFolder, targetVersion.String()); err != nil {
			return err, "Error writing files", 19
		}

		log.Traceln("Moving objects:", app.MoveObjects)
		if err := moveObjects(app.MoveObjects, workingFolder); err != nil {
			return err, "Error moving objects ", 20
		}

		if workingFolder != folderName {
			log.Debugln("Renaming ", workingFolder, " to ", folderName)
			if err := os.Rename(workingFolder, folderName); err != nil {
				return err, "Error renaming folder", 21
			}
		}

		//create/update symlink app-1.0.2 => app ...
		log.Traceln("Handling symlink and restores")

		//Can only restore from previous symlink (targetVersion)....
		if iohelper.FileOrDirExists(symlink) {
			absoluteFolderName, _ := filepath.Abs(folderName)
			evalSymlink, _ := filepath.EvalSymlinks(symlink)
			absoluteSymlink, _ := filepath.Abs(evalSymlink)
			if absoluteFolderName != absoluteSymlink {
				if len(app.RestoreFiles) > 0 {
					//(handles customs/configurations that are overwritten upon upgrade)
					log.Debugln("Restoring ", app.RestoreFiles)
					if err := restoreFiles(app.RestoreFiles, symlink, folderName); err != nil {
						return err, "Error restoring files |", 22
					}
				}
			} else {
				log.Debugln("No version change, skipping restore")
			}

			//Remove old
			err := os.Remove(symlink)
			if err != nil {
				return err, fmt.Sprint("Cannot remove symlink ", symlink, " for update to latest version..."), 0
			}
		}

		//err = os.Mkdir(symlink)
		target := appNameWithVersion + "/"
		log.Debugln("Linking " + symlink + " -> " + target)

		err = junction.Create(target, symlink)
		if err != nil {
			return err, "Error symlink/junction to " + target, 23
		}

		if app.Shortcut != "" {
			const DefaultShortcutsDir = "shortcuts"
			if !iohelper.FileOrDirExists(DefaultShortcutsDir) {
				log.Debugln("Creating shortcutDir ", DefaultShortcutsDir)
				err := os.Mkdir(DefaultShortcutsDir, os.ModePerm)
				if err != nil {
					return err, fmt.Sprint("Cannot create shortcut dir ", DefaultShortcutsDir), 0
				}
			}

			absSymlink, _ := filepath.Abs(symlink)

			pwd, _ := os.Getwd()

			targetForShortcut := path.Join(absSymlink, app.Shortcut)
			if customAppLocationForShortcut != "" {
				targetForShortcut = path.Join(customAppLocationForShortcut, filepath.Base(pwd), symlink, app.Shortcut)
			}

			absShortcutsDir, _ := filepath.Abs(DefaultShortcutsDir)

			icon := ""
			if app.ShortcutIcon != "" {
				icon = path.Join(path.Dir(targetForShortcut), app.ShortcutIcon)
			}

			log.Debugln("Creating shortcut ", app.Shortcut, " -> ", targetForShortcut)
			createShortcut(
				app.Shortcut,
				targetForShortcut,
				"",
				filepath.Dir(targetForShortcut),
				"portable-"+app.Shortcut,
				absShortcutsDir,
				icon)
		}

		return nil, "", 0

	case "status":
		return nil, "", 0
	default:
		return errors.New("Invalid action" + action), "", 81
	}

}

func msiExec(zip string, fullFolderPath string, err error) (error, string, int, bool) {
	// Manually set the arguments since Go escaping does not work with MSI arguments
	argString := fmt.Sprintf(`/a "%v" /qb TARGETDIR="%v"`, zip, fullFolderPath)
	// Build the command
	log.Traceln("msi args: ", argString)
	cmd := exec.Command("msiexec", argString)

	if err = cmd.Run(); err != nil {
		return err, "Error extracting from msi |", 11, true
	}
	return nil, "", 0, false
}

func extractRegex(extension string, zip string, folder string, re *regexp.Regexp) (interface{}, error) {
	switch extension {
	case ".zip":
		return extractZipRegex(zip, folder, re)
	default:
		return "", errors.New("Unsupported extension " + extension)
	}
}

func extractRootFolder(zip string, extension string) (string, error) {
	switch extension {
	case ".zip":
		return extractZipRootFolder(zip)
	default:
		return "", errors.New("Unsupported extension " + extension)
	}
}
