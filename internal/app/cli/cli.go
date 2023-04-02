package cli

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/internal/pkg/installer"
	"github.com/jonathanMelly/nomad/internal/pkg/state"
	versionLib "github.com/jonathanMelly/nomad/pkg/version"
	"golang.org/x/exp/slices"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var versionString string
var versionAdditionalInfos string

func setupBaseConfigAndSettings(githubPat string, versionString string, archivesSubdir string) {
	version, _ := versionLib.FromString(versionString)

	configuration.Version = version
	configuration.Settings.GithubApiKey = githubPat
	configuration.Settings.ArchivesDirectory = archivesSubdir
}

var exeName = filepath.Base(os.Args[0])

func customUsage() {

	printVersion()
	fmt.Printf("Main usage: %s install|update|status [OPTIONS] [...appName]\n\nOPTIONS:\n", exeName)
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("\t", exeName, "i[nstall] rclone")
	fmt.Println("\t", exeName, "u[pgrade] rclone")
	fmt.Println("\t", exeName, "st[atus]")
	fmt.Println("\t", exeName, "-confirm=false install filezilla")
	fmt.Println("\t", exeName, "-verbose u[pdate] git obs")
	fmt.Println("\t", exeName, "v[ersion]")
	fmt.Println("\nList available apps for install:")
	fmt.Println("\t", exeName, "l[ist]")
}

func printVersion() {
	_, err := fmt.Println(exeName, versionString, versionAdditionalInfos)
	if err != nil {
		log.Errorln(err)
	}
}

//goland:noinspection GoSnakeCaseUsage
const (
	EXIT_OK = 0

	EXIT_BAD_USAGE = 4
	EXIT_ACTION    = 41

	EXIT_UNKNOWN_ACTION = 67
	EXIT_NO_VALID_APP   = 68
)

func Main(_embeddedDefs embed.FS, _githubPat string, _version string, _versionExtras string) int {
	versionString = _version //must be done for customUsage
	versionAdditionalInfos = _versionExtras
	flag.Usage = customUsage

	// Overwrite version
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions helper")
	flagDefinitionsDirectory := flag.String("definitions", "app-definitions", "Specify directory for custom app definitions files (relative to current dir and not outside...)")
	flagEnvVarForAppsLocation := flag.String("shortcutLocation", "", "If not empty, will be used for shortcuts generation as base path... (to allow easy switch)")
	flagForceExtract := flag.Bool("force", false, "Remove any similar existing version")
	flagSkipDownload := flag.Bool("skip", true, "Skip download if corresponding archive is already present")
	flagLatestVersion := flag.Bool("latest", true, "If version URL is set, check and use latest version available")
	flagOptimist := flag.Bool("optimist", true, "If true and multiple config given, continue after one failed")
	flagConfirm := flag.Bool("confirm", true, "Asks user to confirm operation")
	flagArchivesSubDir := flag.String("archives", "archives", "Set archives sub dir")
	flagVerbose := flag.Bool("verbose", false, "Verbose mode (mainly for debug)")
	flagVeryVerbose := flag.Bool("vverbose", false, "Very verbose mode (debug)")
	flagRefresh := flag.Bool("refresh", false, "Try to redo files operations, symlinks and shortcuts even with no version bump")

	flag.Parse()

	/*
	   Level 10 = panic, fatal, error, warn, info, debug, & trace
	   Level 5 = panic, fatal, error, warn, info, & debug
	   Level 4 = panic, fatal, error, warn, & info
	   Level 3 = panic, fatal, error, & warn
	   Level 2 = panic, fatal & error
	   Level 1 = panic, fatal
	*/
	if *flagVeryVerbose {
		log.EnableLevelsByNumber(10)
	} else if *flagVerbose {
		log.EnableLevelsByNumber(5)
	} else {
		log.EnableLevelsByNumber(4)
	}

	//Prefix is used to show app name...
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.Traceln("Args", flag.Args())

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(EXIT_BAD_USAGE)
	} else {
		setupBaseConfigAndSettings(_githubPat, _version, *flagArchivesSubDir)
		action := strings.ToLower(flag.Arg(0))

		//VERSION
		if strings.HasPrefix(action, "v") {
			printVersion()
			key := configuration.Settings.GithubApiKey
			log.Debug("Using token ", key[0:int(math.Min(float64(len(key)), 15))], "...\n")
			return EXIT_OK
		} else if strings.HasPrefix(action, "h") /*help*/ {
			customUsage()
			return EXIT_OK
		} else {
			return doAction(
				flagDefinitionsDirectory,
				flagVersion,
				flagLatestVersion,
				action,
				flagForceExtract,
				flagSkipDownload,
				flagEnvVarForAppsLocation,
				flagArchivesSubDir,
				flagConfirm,
				flagOptimist,
				flagRefresh,
				_embeddedDefs)

		}
	}

	return 0

}

func doAction(flagDefinitionsDirectory *string,
	flagVersion *string,
	flagLatestVersion *bool,
	action string,
	flagForceExtract *bool,
	flagSkipDownload *bool,
	flagEnvVarForAppsLocation *string,
	flagArchivesSubDir *string,
	flagConfirm *bool,
	flagOptimist *bool,
	flagRefresh *bool,
	embeddedDefinitions embed.FS) int {
	//LOAD CONFIG/app definitions
	configuration.Load("nomad.toml", *flagDefinitionsDirectory, embeddedDefinitions)

	//sanitize input
	action = strings.ToLower(action)

	//LIST available APPS
	if strings.HasPrefix(action, "l") {
		var result []string
		for app := range configuration.Settings.AppDefinitions {
			result = append(result, app)
		}
		log.Infoln("Available apps:", strings.Join(result, ","))
	} else {

		//Shortcut
		var askedApps []string

		//SELF UPGRADE
		//Alias to update nomad
		if strings.HasPrefix(action, "se") {
			askedApps = []string{"nomad"}
			action = "upgrade"
		} else {
			askedApps = flag.Args()[1:]
		}

		if len(askedApps) > 0 {
			askedApps = state.FilterValidAskedApps(askedApps)
			if len(askedApps) == 0 {
				log.Warnln("No valid app name given")
				return EXIT_NO_VALID_APP
			}
		}

		//Load APPS states and possible actions (upgrade...)
		askedStates := state.LoadAskedAppsInitialStates(askedApps)
		err := state.DeterminePossibleActions(
			askedStates,
			*flagVersion,
			*flagLatestVersion,
			configuration.Settings.GithubApiKey)

		if err != nil {
			log.Errorln("Cannot determine possible actions |", err)
			return EXIT_ACTION
		}

		//STATUS
		if strings.HasPrefix(action, "s") {
			if len(askedStates) == 0 {
				log.Infoln("No app yet installed")
			} else {
				for app, appState := range askedStates {
					log.Info(helper.BuildPrefix(app), appState.StatusMessage())
				}
			}
		} else if slices.IndexFunc([]string{"i", "u"}, func(e string) bool {
			return strings.HasPrefix(action, e)
		}) != -1 {
			//Do the job
			for app, appState := range askedStates {
				log.Debugln("Processing", app)

				exitCode := HandleRun(
					installer.InstallOrUpdate(
						*appState,
						*flagForceExtract,
						*flagSkipDownload,
						*flagEnvVarForAppsLocation,
						*flagArchivesSubDir,
						*flagConfirm,
						*flagRefresh,
					))
				if exitCode != EXIT_OK && !(*flagOptimist) {
					return exitCode
				}
			}

		} else {
			log.Errorln("Unknown action", action)
			return EXIT_UNKNOWN_ACTION
		}
	}
	return EXIT_OK
}

func HandleRun(err error, errorMessage string, exitCode int) int {
	defer log.SetPrefix("")

	if exitCode == 0 {
		log.Debugln("**All seemed to go well ;-)")
	} else {
		if errorMessage != "" {
			if err != nil {
				log.Error(errorMessage, " | ", err.Error())
			} else {
				log.Error(errorMessage)
			}
		} else if err != nil {
			log.Error(err.Error())
		}
	}

	return exitCode
}
