package installer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/lib/bytesize"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func init() {
	// Verbose logging with file name and line number
	//log.SetFlags(log.Lshortfile)
}

// Run will execute commands from a app-definitions file
func Run(action string, configFile string, versionOverwrite string, forceExtract bool, skipDownload bool,
	customAppLocationForShortcut string, archivesSubDir string, useLatestVersion bool, confirm bool) (error error, errorMessage string, exitCode int) {

	if !strings.HasSuffix(configFile, ".json") {
		configFile += ".json"
	}

	// Set the name of the app-definitions file
	configFileName := path.Base(configFile)

	// Output the name of the app-definitions file
	log.Debugln("Loading config: ", configFileName)

	// Load the app-definitions file
	appInfo, err := LoadConfig(configFile)
	if err != nil {
		return err, "Error loading config file " + configFile, 1
	}

	// Compile the regular expressions into one
	re, err := combineRegex(appInfo.ExtractRegExList)
	if err != nil {
		return err, "Error creating regular expression from list", 2
	}

	appNameWithoutVersion := appInfo.ApplicationName[0:strings.LastIndex(appInfo.ApplicationName, "-")]
	log.SetPrefix("|" + appNameWithoutVersion + "| ")

	// Set the folder name
	const DefaultAppPath = "apps"

	//VERSION HANDLING
	var targetVersion *Version = nil
	symlink := ""

	// Overwrite the targetVersion if available
	var versionOverwriteVersion *Version = nil
	if versionOverwrite != "" {
		versionOverwriteVersion, err = fromString(versionOverwrite)
		if err != nil {
			return err, "Bad forced version format: " + appInfo.Version, 31
		}
		log.Debugln("Version flag overwrite:", versionOverwriteVersion)
	}

	configVersion, err := fromString(appInfo.Version)
	if err != nil {
		return err, "Bad version format in config :" + appInfo.Version, 30
	}
	log.Debugln("Version from config: ", configVersion)

	//Current version
	var currentInstalledVersion *Version = nil
	if appInfo.Symlink != "" {
		symlink = path.Join(DefaultAppPath, appInfo.Symlink)
		if FileOrDirExists(symlink) {
			target, _ := os.Readlink(symlink)
			split := strings.Split(filepath.Base(target), "-")
			if len(split) > 1 {
				currentInstalledVersion, err = fromString(split[1])
				log.Debugln("Version installed: ", currentInstalledVersion)
			}
		}
	}

	var latestVersionFromRemote *Version = nil
	// If Version Check parameters are specified
	if useLatestVersion && appInfo.VersionCheck.Url != "" && appInfo.VersionCheck.RegEx != "" {

		// Extract the targetVersion from the webpage
		latestVersionFromRemote, err = extractFromRequest(appInfo.VersionCheck.Url, appInfo.VersionCheck.RegEx)
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
			if latestVersionFromRemote != nil && latestVersionFromRemote.isNewerThan(*configVersion) {
				targetVersion = latestVersionFromRemote
			} else {
				targetVersion = configVersion
			}
		} else /*Already installed*/ {
			if latestVersionFromRemote != nil && latestVersionFromRemote.isNewerThan(*configVersion) && latestVersionFromRemote.isNewerThan(*currentInstalledVersion) {
				targetVersion = latestVersionFromRemote
			} else if configVersion.isNewerThan(*currentInstalledVersion) {
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
			action := "installing"
			log.Infoln(action, "version", targetVersion)
		} else {
			action := "upgrading"
			if currentInstalledVersion.isNewerThan(*targetVersion) {
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

		if !FileOrDirExists(DefaultAppPath) {
			log.Debugln("Creating " + DefaultAppPath + " directory")
			os.Mkdir(DefaultAppPath, os.ModePerm)
		}

		var applicationName = strings.Replace(appInfo.ApplicationName, "{{VERSION}}", targetVersion.String(), -1)
		var folderName = path.Join(DefaultAppPath, applicationName)

		// Set the zip name based off the folder
		// Note: The original file download name will be changed
		var zip = path.Join(DefaultAppPath, archivesSubDir, applicationName+appInfo.DownloadExtension)

		// If the zip file DOES exist on disk
		if FileOrDirExists(zip) {
			// Output the filename of the folder
			log.Debugln("Download Exists:", zip)
		} else {
			zipDir := filepath.Dir(zip)
			if !FileOrDirExists(zipDir) {
				os.MkdirAll(zipDir, os.ModePerm)
			}
		}

		// If SkipDownload is true
		if skipDownload && FileOrDirExists(zip) {
			log.Debugln("Skipping download as ", zip, " archive already exists (use -force to override)")
		} else {
			downloadURL := targetVersion.fillVersionsPlaceholders(appInfo.DownloadUrl)

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
		if FileOrDirExists(folderName) {
			if isDirectory(folderName) {
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

			switch appInfo.DownloadExtension {
			case ".zip", ".7zTODO":
				log.Debugln("ZIP archive")
				// If RemoveRootFolder is set to true
				if appInfo.RemoveRootFolder {
					// If the root folder name is specified
					if len(appInfo.RootFolderName) > 0 {
						workingFolder = appInfo.RootFolderName
					} else { // Else the root folder name is not specified so guess it
						// Return the name of the root folder in the ZIP

						workingFolder, err = extractRootFolder(zip, appInfo.DownloadExtension)
						if err != nil {
							return err, "Error discovering working folder ", 6
						}
					}
				} else {
					rootFolder = workingFolder
				}

				// Extract files based on regular expression
				_, err = extractRegex(appInfo.DownloadExtension, zip, rootFolder, re)
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

				// Build the command
				cmd := exec.Command("msiexec")

				// Manually set the arguments since Go escaping does not work with MSI arguments
				argString := fmt.Sprintf(`/a "%v" /qb TARGETDIR="%v"`, zip, fullFolderPath)
				log.Traceln("msi args: ", argString)
				cmd.SysProcAttr = &syscall.SysProcAttr{
					HideWindow:    false,
					CmdLine:       " " + argString,
					CreationFlags: 0,
				}

				if err = cmd.Run(); err != nil {
					return err, "Error extracting from msi |", 11
				}

				// If RemoveRootFolder is set to true
				if appInfo.RemoveRootFolder {
					// If the root folder name is specified
					if len(appInfo.RootFolderName) > 0 {

						//Get the full path of the folder to set as the root folder
						currentPath := filepath.Join(fullFolderPath, appInfo.RootFolderName)

						// Check to make sure the path is valid
						if currentPath == fullFolderPath {
							return nil, "RootFolderName is invalid:" + appInfo.RootFolderName, 12
						}

						// Copy files based on regular expressions
						if _, err := copyMsiRegex(currentPath, fullFolderPath+"_temp", re); err != nil {
							return err, "Error restore from msi folder |", 13
						}

						// Set the working folder so the rename will work later
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
				return nil, "Download extension not supported:" + appInfo.DownloadExtension, 17
			}
		}

		log.Traceln("Creating folders:", appInfo.CreateFolders)
		if err := createFolders(appInfo.CreateFolders, workingFolder); err != nil {
			return err, "Error creating folders", 18
		}

		log.Traceln("Creating files:", appInfo.CreateFiles)
		if err := writeScripts(appInfo.CreateFiles, workingFolder, targetVersion.String()); err != nil {
			return err, "Error writing files", 19
		}

		log.Traceln("Moving objects:", appInfo.MoveObjects)
		if err := moveObjects(appInfo.MoveObjects, workingFolder); err != nil {
			return err, "Error moving objects ", 20
		}

		if workingFolder != folderName {
			log.Debugln("Renaming ", workingFolder, " to ", folderName)
			if err := os.Rename(workingFolder, folderName); err != nil {
				return err, "Error renaming folder", 21
			}
		}

		if appInfo.Symlink != "" {
			//create/update symlink app-1.0.2 => app ...
			log.Traceln("Handling symlink and restores")

			//Can only restore from previous symlink (targetVersion)....
			if FileOrDirExists(symlink) {
				absoluteFolderName, _ := filepath.Abs(folderName)
				evalSymlink, _ := filepath.EvalSymlinks(symlink)
				absoluteSymlink, _ := filepath.Abs(evalSymlink)
				if absoluteFolderName != absoluteSymlink {
					if len(appInfo.RestoreFiles) > 0 {
						//(handles customs/configurations that are overwritten upon upgrade)
						log.Debugln("Restoring ", appInfo.RestoreFiles)
						if err := restoreFiles(appInfo.RestoreFiles, symlink, folderName); err != nil {
							return err, "Error restoring files |", 22
						}
					}
				} else {
					log.Debugln("No version change, skipping restore")
				}

				//Remove old
				os.Remove(symlink)
			}

			//err = os.Mkdir(symlink)
			target := applicationName + "/"
			log.Debugln("Linking " + symlink + " -> " + target)
			err = os.Symlink(target, symlink)
			if err != nil {
				return err, "Error symlink to " + target, 23
			}

			if appInfo.Shortcut != "" {
				const DefaultShortcutsDir = "shortcuts"
				if !FileOrDirExists(DefaultShortcutsDir) {
					log.Debugln("Creating shortcutDir ", DefaultShortcutsDir)
					os.Mkdir(DefaultShortcutsDir, os.ModePerm)
				}

				absSymlink, _ := filepath.Abs(symlink)

				pwd, _ := os.Getwd()

				targetForShortcut := path.Join(absSymlink, appInfo.Shortcut)
				if customAppLocationForShortcut != "" {
					targetForShortcut = path.Join(customAppLocationForShortcut, filepath.Base(pwd), symlink, appInfo.Shortcut)
				}

				absShortcutsDir, _ := filepath.Abs(DefaultShortcutsDir)

				icon := ""
				if appInfo.ShortcutIcon != "" {
					icon = path.Join(path.Dir(targetForShortcut), appInfo.ShortcutIcon)
				}

				log.Debugln("Creating shortcut ", appInfo.Shortcut, " -> ", targetForShortcut)
				createShortcut(
					appInfo.Shortcut,
					targetForShortcut,
					"",
					filepath.Dir(targetForShortcut),
					"portable-"+appInfo.Shortcut,
					absShortcutsDir,
					icon)
			}
		}

		return nil, "", 0

	case "status":
		return nil, "", 0
	default:
		return errors.New("Invalid action" + action), "", 81
	}

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
