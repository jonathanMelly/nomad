package state

import (
	"fmt"
	"github.com/gologme/log"
	"github.com/gookit/goutil/maputil"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/pkg/version"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	NOT_SET = Status(-1)

	KEEP = Status(0)

	INSTALL = Status(1)

	UPGRADE   = Status(2)
	DOWNGRADE = Status(3)

	// NOT YET IMPLEMENTED UNINSTALL=4
)

type Status int

type AppStates map[string]*AppState

var wg sync.WaitGroup

func NewAppStates() AppStates {
	return AppStates{}
}

type AppState struct {
	Definition           *data.AppDefinition
	SymlinkFound         bool
	CurrentVersion       *version.Version
	TargetVersion        *version.Version
	CurrentVersionFolder string
	Status               Status
}

func FilterValidAskedApps(askedApps []string) (filtered []string) {
	//Check app validity
	for _, askedApp := range askedApps {
		if maputil.HasKey(configuration.Settings.AppDefinitions, askedApp) {
			filtered = append(filtered, askedApp)
		} else {
			log.Warnln("unknown app", askedApp)
		}
	}
	return
}

func LoadAskedAppsInitialStates(askedApps []string) AppStates {

	//Scan installed APPS
	alreadyInstalledStates := ScanCurrentApps(configuration.AppPath)
	log.Debugln("Found", len(alreadyInstalledStates), "installed apps")

	//Handle * asked apps
	askedAppsCount := len(askedApps)
	if askedAppsCount == 0 {
		log.Debugln("Working on all installed apps")
		for app := range alreadyInstalledStates {
			askedApps = append(askedApps, app)
		}
	}
	log.Debugln("Selected apps:", askedApps)

	//Merge installed and not installed states for asked apps
	log.Infoln("Computing status of", len(askedApps), "apps...")
	askedAppsStates := NewAppStates()
	for _, app := range askedApps {
		_state, exist := alreadyInstalledStates[app]
		if exist {
			askedAppsStates[app] = _state
		} else {
			addOrUpdateState("", app, askedAppsStates, nil, false)
		}
	}

	//Special nomad (always get version from binary)
	nomadState, found := askedAppsStates["nomad"]
	if found {
		log.Traceln("Getting nomad version from binary")
		nomadState.CurrentVersion = configuration.Version
	}

	return askedAppsStates

}

// ScanCurrentApps Look for installed apps matching available definitions
func ScanCurrentApps(baseDirectory string) AppStates {
	log.Traceln("Searching for currently installed apps in", baseDirectory)

	installedApps := NewAppStates()

	if helper.FileOrDirExists(baseDirectory) {
		files, err := os.ReadDir(baseDirectory)
		if err != nil {
			log.Fatal(err)
		}

		alreadyAnalyzedThroughSymlinks := map[string]bool{}
		for _, f := range files {
			var targetDirectory string
			fullPath := filepath.Join(baseDirectory, f.Name())
			//First dereference link if it's one
			isLink := helper.IsSymlink(fullPath)
			if isLink {
				log.Traceln("Found link", fullPath)
				link, err := os.Readlink(fullPath)
				if err != nil {
					log.Errorln("Cannot read link", fullPath, "|", err)
					continue
				} else {
					linkInfo, err := os.Stat(link)
					if err != nil {
						log.Errorln("Cannot stat", link, "|", err)
						continue
					} else if linkInfo.IsDir() {
						log.Traceln("Link", fullPath, "is pointing to valid directory", link)
						cwd, err := os.Getwd()

						//guarantee that path is relative (with win junction it may be abs)
						relPath, err := filepath.Rel(filepath.Join(cwd, baseDirectory), link)
						if err != nil {
							log.Errorln("Cannot get relative path, base=", cwd, "target=", link, "|", err)
							continue
						} else {
							targetDirectory = relPath
							alreadyAnalyzedThroughSymlinks[targetDirectory] = true
						}
					}
				}
			} else if isDir := f.IsDir(); isDir {
				if _, already := alreadyAnalyzedThroughSymlinks[f.Name()]; !already {
					targetDirectory = f.Name()
				} else {
					log.Traceln("Discarding", f.Name(), "folder as already scanned through symlink")
				}
			}

			if targetDirectory != "" {
				analyzeEntry(baseDirectory, targetDirectory, installedApps, isLink)
			}
		}
	} else {
		log.Debugln("Directory", baseDirectory, "does not exist (no apps yet installed)")
	}

	return installedApps
}

func analyzeEntry(rootPath string, appDirectory string, states AppStates, isSymlink bool) {
	fullPath := filepath.Join(rootPath, appDirectory)
	log.Traceln("Analyzing", fullPath, "(from symlink:", isSymlink, ")")

	const VersionSeparator = "-"
	lastSeparatorPosition := strings.LastIndex(appDirectory, VersionSeparator)
	if lastSeparatorPosition >= 0 {
		guessedApp := appDirectory[:lastSeparatorPosition]
		guessedVersionString := appDirectory[lastSeparatorPosition+1:]
		log.Traceln("Guessed app", guessedApp, "with version", guessedVersionString)

		guessedVersion, err := version.FromString(guessedVersionString)
		if err != nil {
			log.Errorln("Cannot get version of", guessedApp, "->skipping")
		} else {
			//Symlink is "MASTER"
			if isSymlink {
				log.Traceln("Setting current app", guessedApp, "to", guessedVersion, " (symlink found)")
				addOrUpdateState(fullPath, guessedApp, states, guessedVersion, isSymlink)
			} else {
				identifiedApp, alreadyFound := states[guessedApp]

				//Is it a better candidate for current active version ? (remember, symlink is MASTER)
				if alreadyFound {
					if !identifiedApp.SymlinkFound && !identifiedApp.CurrentVersion.IsNewerThan(guessedVersion) {
						addOrUpdateState(fullPath, guessedApp, states, guessedVersion, isSymlink)
					} else {
						log.Traceln("Discarding existing app", guessedApp, "with version", guessedVersion, "as current version candidate (symlinked or newer with version", identifiedApp.CurrentVersion, "already found)")
					}
				} else {
					addOrUpdateState(fullPath, guessedApp, states, guessedVersion, isSymlink)
				}
			}
		}
	} else {
		log.Traceln("No", VersionSeparator, "found in", fullPath, ", discarding entry")
	}
}

func addOrUpdateState(appPath string, guessedApp string, states AppStates, currentVersion *version.Version, isSymlink bool) {
	knownAppDef, knownApp := configuration.Settings.AppDefinitions[guessedApp]
	if knownApp {
		updatedState := buildState(appPath, *knownAppDef, isSymlink, currentVersion)
		states[guessedApp] = &updatedState
	} else {
		log.Traceln("Unknown app", guessedApp)
	}
}

func buildState(appPath string, knownAppDef data.AppDefinition, isSymlink bool, currentVersion *version.Version) AppState {
	return AppState{
		Definition:           &knownAppDef,
		SymlinkFound:         isSymlink,
		CurrentVersion:       currentVersion,
		CurrentVersionFolder: appPath,
		TargetVersion:        nil,
	}
}

func DeterminePossibleActions(
	apps AppStates,
	forceVersion string,
	useLatestVersion bool, apiKey string) error {

	//Load Versioning info
	forcedVersion, err := validateForcedVersionIfNeeded(forceVersion)
	if err != nil {
		log.Errorln("Bad forced version format:", forceVersion)
		return err
	}

	defer log.SetPrefix("")
	wg.Add(len(apps))
	for appName, state := range apps {
		go computeState(appName, state, useLatestVersion, apiKey, forcedVersion)
	}
	wg.Wait()

	return nil

}

func computeState(appName string, state *AppState, useLatestVersion bool, apiKey string, forcedVersion *version.Version) {
	defer wg.Done()
	log.SetPrefix(fmt.Sprint("|", appName, "| "))

	var configVersion *version.Version = nil
	if state.Definition.Version != "" {
		var err error
		configVersion, err = version.FromString(state.Definition.Version)
		if err != nil {
			log.Errorln("Bad version format in config : ", state.Definition.Version, "|", err)
		}
	}

	log.Debugln("Version from config: ", configVersion)

	//Current version
	currentInstalledVersion := state.CurrentVersion
	log.Debugln("Version installed: ", currentInstalledVersion)

	var latestVersionFromRemote *version.Version = nil
	// If Version Check parameters are specified
	if useLatestVersion && state.Definition.VersionCheck.Url != "" && state.Definition.VersionCheck.RegEx != "" {
		url, requestBody := state.Definition.VersionCheck.BuildRequest()

		// Extract the targetVersion from the webpage
		var err error
		latestVersionFromRemote, err =
			helper.GetVersion(url, state.Definition, apiKey, requestBody)
		if err != nil {
			log.Errorln("Error retrieving last version from remote", err)
		}
	}
	log.Debugln("Version from remote: ", latestVersionFromRemote)

	var targetVersion *version.Version
	if forcedVersion != nil {
		targetVersion = forcedVersion
	} else {
		//not yet installed
		if currentInstalledVersion == nil {
			if latestVersionFromRemote != nil && latestVersionFromRemote.IsNewerThan(configVersion) {
				targetVersion = latestVersionFromRemote
			} else {
				targetVersion = configVersion
			}
		} else /*Already installed*/ {
			if latestVersionFromRemote != nil && latestVersionFromRemote.IsNewerThan(configVersion) && latestVersionFromRemote.IsNewerThan(currentInstalledVersion) {
				targetVersion = latestVersionFromRemote
			} else if configVersion != nil && configVersion.IsNewerThan(currentInstalledVersion) {
				log.Debugln("Config ", configVersion, " is newer than currentInstalled", currentInstalledVersion)
				targetVersion = configVersion
			} else {
				targetVersion = currentInstalledVersion
			}
		}
	}
	log.Debugln("target version", targetVersion)
	state.TargetVersion = targetVersion
	state.computeStatus()
}

func (state *AppState) StatusMessage() string {
	if state.Status == NOT_SET {
		state.computeStatus()
	}
	switch state.Status {
	case KEEP:
		return fmt.Sprint("installed version ", state.CurrentVersion, " is already up to date")
	case INSTALL:
		return fmt.Sprint("not installed >> will install version ", state.TargetVersion)
	case UPGRADE:
		return fmt.Sprint("upgrading version from ", state.CurrentVersion, " >> ", state.TargetVersion)
	case DOWNGRADE:
		return fmt.Sprint("downgrading version from ", state.CurrentVersion, " >> ", state.TargetVersion)
	default:
		return ""
	}
}

func (state *AppState) SuccessMessage() string {
	if state.Status == NOT_SET {
		state.computeStatus()
	}
	switch state.Status {
	case KEEP:
		return fmt.Sprint("successfully kept at version ", state.CurrentVersion)
	case INSTALL:
		return fmt.Sprint("version ", state.TargetVersion, " successfully installed ")
	case UPGRADE:
		return fmt.Sprint("successfully upgraded from ", state.CurrentVersion, " to ", state.TargetVersion)
	case DOWNGRADE:
		return fmt.Sprint("successfully downgraded from ", state.CurrentVersion, " to ", state.TargetVersion)
	default:
		return ""
	}
}

func (state *AppState) computeStatus() {
	if state.CurrentVersion == nil {
		state.Status = INSTALL
	} else {
		if state.TargetVersion.IsNewerThan(state.CurrentVersion) {
			state.Status = UPGRADE
		} else if state.CurrentVersion.IsNewerThan(state.TargetVersion) {
			state.Status = DOWNGRADE
		} else {
			state.Status = KEEP
		}
	}
}

func validateForcedVersionIfNeeded(forceVersion string) (*version.Version, error) {
	if forceVersion != "" {
		versionOverwriteVersion, err := version.FromString(forceVersion)
		if err != nil {
			log.Errorln("Bad forced version format: ", forceVersion, "|", err)
			return nil, err
		} else {
			log.Debugln("Version from cmd:", versionOverwriteVersion)
			return versionOverwriteVersion, nil
		}

	}
	return nil, nil
}
