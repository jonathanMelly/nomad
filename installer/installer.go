package installer

import (
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
func Run(configFile string, versionOverwrite string, forceExtract bool, skipDownload bool, envVarForAppsLocation string, archivesSubDir string) {

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
		log.Println(err)
		unifiedExit(1)
	}

	// Compile the regular expressions into one
	re, err := combineRegex(appInfo.ExtractRegExList)
	if err != nil {
		log.Println("Error creating regular expression from list |", err)
		unifiedExit(1)
	}

	// Set the application version
	version := appInfo.Version

	// Overwrite the version if available
	if versionOverwrite != "" {
		version = versionOverwrite
	}

	// If Version Check parameters are specified
	if appInfo.VersionCheck.Url != "" && appInfo.VersionCheck.RegEx != "" {
		re := strings.Replace(appInfo.VersionCheck.RegEx, "{{VERSION}}", `(\d+(?:\.\d+)*)`, 1)
		// Extract the version from the webpage
		newVersion, err := extractFromRequest(appInfo.VersionCheck.Url, re)
		if err != nil {
			log.Println("Error retrieving page |", err)
			unifiedExit(1)
		}

		if newVersion != version {
			log.Println("Newer version available: " + newVersion)
		}

		if appInfo.VersionCheck.UseLatestVersion {
			version = newVersion
		}

		log.Println("Using version: " + version)
	}

	// Set the folder name
	const DefaultAppPath = "apps"
	const DefaultShortcutPath = `%userprofile%\portable-app-shortcuts`

	if !isExist(DefaultAppPath) {
		log.Println("Creating " + DefaultAppPath + " directory")
		os.Mkdir(DefaultAppPath, os.ModePerm)
	}

	var applicationName = strings.Replace(appInfo.ApplicationName, "{{VERSION}}", version, -1)
	var folderName = path.Join(DefaultAppPath, applicationName)

	// Set the zip name based off the folder
	// Note: The original file download name will be changed
	var zip = path.Join(DefaultAppPath, archivesSubDir, applicationName+appInfo.DownloadExtension)

	// If the zip file DOES exist on disk
	if isExist(zip) {
		// Output the filename of the folder
		log.Println("Download Exists:", zip)
	} else {
		zipDir := filepath.Dir(zip)
		if !isExist(zipDir) {
			os.MkdirAll(zipDir, os.ModePerm)
		}
	}

	// If SkipDownload is true
	if skipDownload && isExist(zip) {
		log.Println("Skipping download")
	} else {
		log.Println("Will Download:", folderName)

		versionRegex := regexp.MustCompile(`(\d+)(?:\.(\d+)*)`)
		versionMatch := versionRegex.FindStringSubmatch(version)

		versionMajor := versionMatch[1]
		versionMinor := "0"
		if len(versionMatch) > 2 {
			versionMinor = versionMatch[2]
		}
		versionBugfix := "0"
		if len(versionMatch) > 3 {
			versionBugfix = versionMatch[3]
		}
		log.Println("Version details: Major:" + versionMajor + "|Minor:" + versionMinor + "|BugFix:" + versionBugfix)

		var versionReplaces = map[string]string{
			"VERSION":        version,
			"VERSION_NO_DOT": strings.ReplaceAll(version, ".", ""),
			"V_MINOR":        versionMinor,
			"V_MAJOR":        versionMajor,
			"V_BUGFIX":       versionBugfix,
		}
		downloadURL := appInfo.DownloadUrl
		for source, replacement := range versionReplaces {
			downloadURL = strings.Replace(downloadURL, "{{"+source+"}}", replacement, -1)
		}

		log.Println("Downloading from:", downloadURL)
		log.Println("Downloading to:", zip)

		size, err := downloadFile(downloadURL, zip)
		if err != nil {
			log.Println("Error download file |", err)
			unifiedExit(1)
		}
		log.Println("Download Size:", bytesize.ByteSize(size))
	}

	extract := true
	// If the folder exists
	if isExist(folderName) {
		if isDirectory(folderName) {
			if forceExtract {
				log.Println("Removing old folder:", folderName)
				err = os.RemoveAll(folderName)
				if err != nil {
					log.Println("Error removing working folder |", err)
					unifiedExit(1)
				}

			} else {
				log.Println("Folder already exists:", folderName)
				log.Println("*** No change")
				extract = false
			}
		} else {
			log.Println("/!\\WARNINIG, filename (not directory) " + folderName + " already existing, please check")
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
						log.Println("Error discovering working folder |", err)
						unifiedExit(1)
					}
				}
			} else {
				rootFolder = workingFolder
			}

			// Extract files based on regular expression
			_, err = extractRegex(appInfo.DownloadExtension, zip, rootFolder, re)
			if err != nil {
				log.Println("Error extracting from zip |", err)
				unifiedExit(1)
			}
		case ".msi":
			// Make the folder
			err = os.Mkdir(folderName, os.ModePerm)
			if err != nil {
				log.Println("Error making folder |", err)
				unifiedExit(1)
			}

			// Get the full folder path
			fullFolderPath, err := filepath.Abs(folderName)
			if err != nil {
				log.Println("Error getting folder full path |", err)
				unifiedExit(1)
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

			err = cmd.Run()
			if err != nil {
				log.Println("Error extracting from msi |", err)
				unifiedExit(1)
			}

			// If RemoveRootFolder is set to true
			if appInfo.RemoveRootFolder {
				// If the root folder name is specified
				if len(appInfo.RootFolderName) > 0 {

					//Get the full path of the folder to set as the root folder
					currentPath := filepath.Join(fullFolderPath, appInfo.RootFolderName)

					// Check to make sure the path is valid
					if currentPath == fullFolderPath {
						log.Println("RootFolderName is invalid:", appInfo.RootFolderName)
						unifiedExit(1)
					}

					// Copy files based on regular expressions
					_, err = copyMsiRegex(currentPath, fullFolderPath+"_temp", re)
					if err != nil {
						log.Println("Error restore from msi folder |", err)
						unifiedExit(1)
					}

					// Set the working folder so the rename will work later
					workingFolder = fullFolderPath + "_temp"

					// Remove the original full folder path
					err = os.RemoveAll(fullFolderPath)
					if err != nil {
						log.Println("Error removing MSI folder:", currentPath)
						unifiedExit(1)
					}

				} else { // Else the root folder name is not specified
					log.Println("The string, RemoveRootName, is required for MSIs")
					unifiedExit(1)
				}
			} else {
				log.Println("The boolean, RemoveRootFolder, is required for MSIs")
				unifiedExit(1)
			}
		default:
			log.Println("Download extension not supported:", appInfo.DownloadExtension)
			unifiedExit(1)
		}
	}

	log.Println("Creating folders")
	err = createFolders(appInfo.CreateFolders, workingFolder)
	if err != nil {
		log.Println("Error creating folders |", err)
		unifiedExit(1)
	}

	log.Println("Creating files")
	err = writeScripts(appInfo.CreateFiles, workingFolder, version)
	if err != nil {
		log.Println("Error writing files |", err)
		unifiedExit(1)
	}

	log.Println("Moving objects")
	err = moveObjects(appInfo.MoveObjects, workingFolder)
	if err != nil {
		log.Println("Error moving objects |", err)
		unifiedExit(1)
	}

	if workingFolder != folderName {
		log.Println("Renaming folder to:", folderName)
		err = os.Rename(workingFolder, folderName)
		if err != nil {
			log.Println("Error renaming folder |", err)
			unifiedExit(1)
		}
	}

	if appInfo.Symlink != "" {
		//create/update symlink app-1.0.2 => app ...
		log.Println("Handling symlink and restores")
		symlink := path.Join(DefaultAppPath, appInfo.Symlink)

		//Can only restore from previous symlink (version)....
		if isExist(symlink) {
			absoluteFolderName, _ := filepath.Abs(folderName)
			evalSymlink, _ := filepath.EvalSymlinks(symlink)
			absoluteSymlink, _ := filepath.Abs(evalSymlink)
			if absoluteFolderName != absoluteSymlink {
				if len(appInfo.RestoreFiles) > 0 {
					//(handles customs/configurations that are overwritten upon upgrade)

					err = restoreFiles(appInfo.RestoreFiles, symlink, folderName)
					if err != nil {
						log.Println("Error restoring files |", err)
						unifiedExit(1)
					}
				}
			} else {
				log.Println("=>Same app version, skipping restore")
			}

			//Remove old
			os.Remove(symlink)
		}

		//err = os.Mkdir(symlink)
		target := applicationName + "/"
		log.Println("Linking " + symlink + "->" + target)
		err = os.Symlink(target, symlink)
		if err != nil {
			log.Println("Error symlink to "+target+" |", err)
			unifiedExit(1)
		}

		if appInfo.Shortcut != "" {
			const DefaultShortcutsDir = "shortcuts"
			if !isExist(DefaultShortcutsDir) {
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

	//unifiedExit(0)
	log.Println("*** Success")
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
