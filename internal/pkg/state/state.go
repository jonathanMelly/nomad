package state

import (
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/pkg/version"
	"os"
	"path/filepath"
	"strings"
)

// ScanCurrentApps Look for installed apps matching available definitions
func ScanCurrentApps(directory string) *data.AppStates {
	installedApps := data.NewAppStates()

	if helper.FileOrDirExists(directory) {
		files, err := os.ReadDir(directory)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if f.IsDir() {
				subDirectory := filepath.Join(directory, f.Name())
				log.Traceln("Inspecting dir", subDirectory)
				analyzeEntry(directory, f.Name(), installedApps, false)
			}
		}
	} else {
		log.Debugln("Directory", directory, "does not exist (no apps yet installed)")
	}

	return installedApps
}

func AppendInstallableApps(candidates []string, bucketToPatch *data.AppStates) {
	for _, app := range candidates {
		_, installed := bucketToPatch.States[app]
		if !installed {
			updateAppState(app, bucketToPatch, nil, false)
		}
	}
}

func analyzeEntry(rootPath string, appDirectory string, bucket *data.AppStates, isSymlink bool) {
	fullPath := filepath.Join(rootPath, appDirectory)
	guessedApp, guessedVersionString, dashFound := strings.Cut(appDirectory, "-")
	if dashFound {
		guessedVersion, err := version.FromString(guessedVersionString)
		if err != nil {
			log.Errorln("Cannot get version of", guessedApp, "->skipping")
		} else {
			//Symlink is "MASTER"
			if isSymlink {
				updateAppState(guessedApp, bucket, guessedVersion, isSymlink)
			} else {
				identifiedApp, alreadyFound := bucket.States[guessedApp]

				if alreadyFound {
					if !identifiedApp.SymlinkFound && !identifiedApp.CurrentVersion.IsNewerThan(*guessedVersion) {
						updateAppState(guessedApp, bucket, guessedVersion, isSymlink)
					} else {
						log.Debugln("Discarding older version", guessedApp, guessedVersion)
					}
				} else {
					updateAppState(guessedApp, bucket, guessedVersion, isSymlink)
				}
			}
		}
	} else if helper.IsSymlink(fullPath) {
		log.Debugln("Analyzing symlink", fullPath)
		target, err := os.Readlink(fullPath)
		if err != nil {
			log.Errorln("Cannot read link", fullPath, err)
		} else {
			analyzeEntry(rootPath, target, bucket, true)
		}
	}
}

func updateAppState(guessedApp string, bucket *data.AppStates, currentVersion *version.Version, isSymlink bool) {
	knownAppDef, knownApp := configuration.Settings.AppDefinitions[guessedApp]
	if knownApp {
		updatedState := buildState(*knownAppDef, isSymlink, currentVersion)
		bucket.States[guessedApp] = &updatedState
	} else {
		log.Warnln("Unknown app", guessedApp)
	}
}

func buildState(knownAppDef data.AppDefinition, isSymlink bool, currentVersion *version.Version) data.AppState {
	return data.AppState{
		Definition:     knownAppDef,
		SymlinkFound:   isSymlink,
		CurrentVersion: currentVersion,
		TargetVersion:  nil,
	}
}

func DetermineActionToBePerformed(
	bucket data.AppStates,
	forceVersion *version.Version,
	useLatestVersion bool, apiKey string) {

	defer log.SetPrefix("")

	for appName, state := range bucket.States {
		log.SetPrefix(fmt.Sprint("|", appName, "| "))

		configVersion, err := version.FromString(state.Definition.Version)
		if err != nil {
			log.Errorln("Bad version format in config : ", state.Definition.Version, "|", err)
		}
		log.Debugln("Version from config: ", configVersion)

		//Current version
		currentInstalledVersion := state.CurrentVersion
		log.Debugln("Version installed: ", currentInstalledVersion)

		var latestVersionFromRemote *version.Version = nil
		// If Version Check parameters are specified
		if useLatestVersion && state.Definition.VersionCheck.Url != "" && state.Definition.VersionCheck.RegEx != "" {

			// Extract the targetVersion from the webpage
			latestVersionFromRemote, err = helper.ExtractFromRequest(state.Definition.VersionCheck.Url, state.Definition.VersionCheck.RegEx, apiKey)
			if err != nil {
				log.Errorln("Error retrieving last version from remote", err)
			}
			log.Debugln("Version from remote: ", latestVersionFromRemote)

		}

		var targetVersion *version.Version
		if forceVersion != nil {
			targetVersion = forceVersion
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
		state.TargetVersion = targetVersion

		message := BuildActionMessage(currentInstalledVersion, targetVersion)
		state.ActionMessage = message

	}

}

func BuildActionMessage(currentInstalledVersion *version.Version, targetVersion *version.Version) string {
	var message string
	currentInstalledVersionString := "not installed"
	if currentInstalledVersion != nil {
		currentInstalledVersionString = currentInstalledVersion.String()
	}
	if currentInstalledVersionString != targetVersion.String() {
		if currentInstalledVersion == nil {
			action := "not installed >> will install"
			message = fmt.Sprint(action, " version ", targetVersion)
		} else {
			action := "upgrading"
			if currentInstalledVersion.IsNewerThan(*targetVersion) {
				action = "downgrading"
			}
			message = fmt.Sprint(action, " version from ", currentInstalledVersionString, " >> ", targetVersion)
		}

	} else {
		message = fmt.Sprint("installed version ", targetVersion, " is already up to date")
	}
	return message
}
