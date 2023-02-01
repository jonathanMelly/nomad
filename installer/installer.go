package installer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jonathanMelly/portable-app-installer/lib/bytesize"
	"log"
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
func Run(configFile string, versionOverwrite string, forceExtract bool, skipDownload bool,
	envVarForAppsLocation string, archivesSubDir string, useLatestVersion bool, confirm bool) (error error, errorMessage string, exitCode int) {

	if !strings.HasSuffix(configFile, ".json") {
		configFile += ".json"
	}

	// Set the name of the app-definitions file
	configFileName := path.Base(configFile)

	// Output the name of the app-definitions file
	log.Println("*** Loading: " + configFileName)

	// Load the app-definitions file
	appInfo, err := LoadConfig(configFile)
	if err != nil {
		return err, "", 1
	}

	// Compile the regular expressions into one
	re, err := combineRegex(appInfo.ExtractRegExList)
	if err != nil {
		return err, "Error creating regular expression from list", 2
	}

	// Set the folder name
	const DefaultAppPath = "apps"
	const DefaultShortcutPath = `%userprofile%\portable-app-shortcuts`

	//VERSION HANDLING
	// Set the application targetVersion
	targetVersion := appInfo.Version

	// Overwrite the targetVersion if available
	if versionOverwrite != "" {
		targetVersion = versionOverwrite
	}

	//Current version
	currentVersion := ""
	symlink := ""
	if appInfo.Symlink != "" {
		symlink = path.Join(DefaultAppPath, appInfo.Symlink)
		if fileOrDirExists(symlink) {
			target, _ := os.Readlink(symlink)
			split := strings.Split(filepath.Base(target), "-")
			if len(split) > 1 {
				currentVersion = split[1]
				log.Println("Found current version : ", currentVersion)
				targetVersion = currentVersion
			}
		}
	}

	const VersionRegex = `(\d+)(?:\.(\d+)*)(-(?:rc|beta)\.\d)?`

	// If Version Check parameters are specified
	if appInfo.VersionCheck.Url != "" && appInfo.VersionCheck.RegEx != "" {

		re := strings.Replace(appInfo.VersionCheck.RegEx, "{{VERSION}}", VersionRegex, 1)
		// Extract the targetVersion from the webpage
		newVersion, err := extractFromRequest(appInfo.VersionCheck.Url, re)
		if err != nil {
			return error, "Error retrieving page", 3
		}

		//TODO check if really newer (not different)...
		if newVersion != targetVersion {
			log.Println("Newer version available: " + newVersion)
			if appInfo.VersionCheck.UseLatestVersion || useLatestVersion {

				targetVersion = newVersion
			}
		}
	}

	log.Println("Target version to install: " + targetVersion)

	if confirm {
		scanner := bufio.NewScanner(os.Stdin)
		log.Print("Proceed to install [Y,n] ?")
		scanner.Scan()
		answer := scanner.Text()
		if !(answer == "" || strings.ToLower(answer) == "y") {
			return nil, "Install aborted by user", 10
		}
	}

	if !fileOrDirExists(DefaultAppPath) {
		log.Println("Creating " + DefaultAppPath + " directory")
		os.Mkdir(DefaultAppPath, os.ModePerm)
	}

	var applicationName = strings.Replace(appInfo.ApplicationName, "{{VERSION}}", targetVersion, -1)
	var folderName = path.Join(DefaultAppPath, applicationName)

	// Set the zip name based off the folder
	// Note: The original file download name will be changed
	var zip = path.Join(DefaultAppPath, archivesSubDir, applicationName+appInfo.DownloadExtension)

	// If the zip file DOES exist on disk
	if fileOrDirExists(zip) {
		// Output the filename of the folder
		log.Println("Download Exists:", zip)
	} else {
		zipDir := filepath.Dir(zip)
		if !fileOrDirExists(zipDir) {
			os.MkdirAll(zipDir, os.ModePerm)
		}
	}

	versionRegex := regexp.MustCompile(VersionRegex)
	versionMatch := versionRegex.FindStringSubmatch(targetVersion)

	versionMajor := versionMatch[1]
	versionMinor := "0"
	if len(versionMatch) > 2 {
		versionMinor = versionMatch[2]
	}
	versionBugfix := "0"
	if len(versionMatch) > 3 {
		versionBugfix = versionMatch[3]
	}
	versionSuffix := ""
	if len(versionMatch) > 4 {
		versionBugfix = versionMatch[4]
	}
	log.Println("Version details: Major:", versionMajor, "|Minor:", versionMinor, "|BugFix:", versionBugfix)

	var versionReplaces = map[string]string{
		"VERSION":        targetVersion,
		"VERSION_NO_DOT": strings.ReplaceAll(targetVersion, ".", ""),
		"V_MINOR":        versionMinor,
		"V_MAJOR":        versionMajor,
		"V_BUGFIX":       versionBugfix,
		"V_SUFFIX":       versionSuffix,
	}

	// If SkipDownload is true
	if skipDownload && fileOrDirExists(zip) {
		log.Println("Skipping download")
	} else {
		log.Println("Will Download:", folderName)

		downloadURL := appInfo.DownloadUrl
		for source, replacement := range versionReplaces {
			downloadURL = strings.Replace(downloadURL, "{{"+source+"}}", replacement, -1)
		}

		log.Println("Downloading from:", downloadURL)
		log.Println("Downloading to:", zip)

		size, err := downloadFile(downloadURL, zip)
		if err != nil {
			return err, "Error download file", 4
		}
		log.Println("Download Size:", bytesize.ByteSize(size))
	}

	extract := true
	// If the folder exists
	if fileOrDirExists(folderName) {
		if isDirectory(folderName) {
			if forceExtract {
				log.Println("Removing old folder:", folderName)
				err = os.RemoveAll(folderName)
				if err != nil {
					return err, "Error removing working folder |", 5
				}

			} else {
				log.Println("Folder already exists:", folderName)
				log.Println("*** No change")
				extract = false
			}
		} else {
			log.Println("/!\\WARNINIG, filename (not directory) ", folderName, " already existing, please check")
			extract = false
		}

	}

	// Working folder is the root folder where the files will be extracted
	workingFolder := folderName

	// Root folder is directory relative to the current directory where the files
	// will be extracted to
	rootFolder := ""

	if extract {
		log.Println("Extracting files")

		switch appInfo.DownloadExtension {
		case ".zip", ".7zTODO":
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
			log.Println(argString)
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

	log.Println("Creating folders")

	if err := createFolders(appInfo.CreateFolders, workingFolder); err != nil {
		return err, "Error creating folders", 18
	}

	log.Println("Creating files")

	if err := writeScripts(appInfo.CreateFiles, workingFolder, targetVersion); err != nil {
		return err, "Error writing files", 19
	}

	log.Println("Moving objects")

	if err := moveObjects(appInfo.MoveObjects, workingFolder); err != nil {
		return err, "Error moving objects ", 20
	}

	if workingFolder != folderName {
		log.Println("Renaming folder to:", folderName)
		if err := os.Rename(workingFolder, folderName); err != nil {
			return err, "Error renaming folder", 21
		}
	}

	if appInfo.Symlink != "" {
		//create/update symlink app-1.0.2 => app ...
		log.Println("Handling symlink and restores")

		//Can only restore from previous symlink (targetVersion)....
		if fileOrDirExists(symlink) {
			absoluteFolderName, _ := filepath.Abs(folderName)
			evalSymlink, _ := filepath.EvalSymlinks(symlink)
			absoluteSymlink, _ := filepath.Abs(evalSymlink)
			if absoluteFolderName != absoluteSymlink {
				if len(appInfo.RestoreFiles) > 0 {
					//(handles customs/configurations that are overwritten upon upgrade)

					if err := restoreFiles(appInfo.RestoreFiles, symlink, folderName); err != nil {
						return err, "Error restoring files |", 22
					}
				}
			} else {
				log.Println("=>Same app targetVersion, skipping restore")
			}

			//Remove old
			os.Remove(symlink)
		}

		//err = os.Mkdir(symlink)
		target := applicationName + "/"
		log.Println("Linking " + symlink + "->" + target)
		err = os.Symlink(target, symlink)
		if err != nil {
			return err, "Error symlink to " + target, 23
		}

		if appInfo.Shortcut != "" {
			const DefaultShortcutsDir = "shortcuts"
			if !fileOrDirExists(DefaultShortcutsDir) {
				os.Mkdir(DefaultShortcutsDir, os.ModePerm)
			}

			absSymlink, _ := filepath.Abs(symlink)

			pwd, _ := os.Getwd()

			target := path.Join(absSymlink, appInfo.Shortcut)
			if envVarForAppsLocation != "" {
				target = path.Join(envVarForAppsLocation, filepath.Base(pwd), symlink, appInfo.Shortcut)
			}

			absShortcutsDir, _ := filepath.Abs(DefaultShortcutsDir)

			icon := ""
			if appInfo.ShortcutIcon != "" {
				icon = path.Join(path.Dir(target), appInfo.ShortcutIcon)
			}

			createShortcut(
				appInfo.Shortcut,
				target,
				"",
				filepath.Dir(target),
				"portable-"+appInfo.Shortcut,
				absShortcutsDir,
				icon)
		}
	}

	return nil, "", 0
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
