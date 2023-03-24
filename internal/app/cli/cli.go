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
	"math"
	"os"
	"path/filepath"
	"strings"
)

var Version = "unreleased"

var EmbeddedDefinitions embed.FS

var exeName = filepath.Base(os.Args[0])

func customUsage() {

	printVersion()
	fmt.Printf("Main usage: %s install|update|status [OPTIONS] [...appName]\n\nOPTIONS:\n", exeName)
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("\t", exeName, "install rclone")
	fmt.Println("\t", exeName, "status")
	fmt.Println("\t", exeName, "-confirm=false install filezilla")
	fmt.Println("\t", exeName, "-verbose update git obs")
	fmt.Println("\t", exeName, "version")
	fmt.Println("\nList available apps for install:")
	fmt.Println("\t", exeName, "list")
}

func printVersion() {
	_, err := fmt.Println(exeName, Version)
	if err != nil {
		log.Errorln(err)
	}
}

//goland:noinspection GoSnakeCaseUsage
const (
	EXIT_OK = 0

	EXIT_BAD_USAGE          = 4
	EXIT_BAD_FORCED_VERSION = 41

	EXIT_UNKNOWN_ACTION = 67
)

func Main() int {

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
	configuration.Settings.ArchivesDirectory = *flagArchivesSubDir

	//Prefix is used to show app name...
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.Traceln("Args", flag.Args())

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(EXIT_BAD_USAGE)
	} else {
		action := flag.Arg(0)

		if action == "version" {
			printVersion()
			key := configuration.Settings.GithubApiKey
			log.Debug("Using token ", key[0:int(math.Min(float64(len(key)), 15))], "...\n")
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
				flagOptimist)

		}
	}

	return 0

}

func doAction(
	flagDefinitionsDirectory *string,
	flagVersion *string,
	flagLatestVersion *bool,
	action string,
	flagForceExtract *bool,
	flagSkipDownload *bool,
	flagEnvVarForAppsLocation *string,
	flagArchivesSubDir *string,
	flagConfirm *bool,
	flagOptimist *bool,
) int {
	//LOAD CONFIG
	configuration.Load("nomad.toml", *flagDefinitionsDirectory, EmbeddedDefinitions)

	//Load APPS states
	askedStates := state.LoadAskedAppsInitialStates(flag.Args()[1:]...)

	err := state.DeterminePossibleActions(
		askedStates,
		*flagVersion,
		*flagLatestVersion,
		configuration.Settings.GithubApiKey)

	if err != nil {
		log.Errorln("Bad forced version", *flagVersion, "|", err)
		return EXIT_BAD_FORCED_VERSION
	}

	switch action {
	case "status":
		if len(askedStates) == 0 {
			log.Infoln("No app yet alreadyInstalled")
		} else {
			log.Infoln("Computing status of", len(askedStates), "apps...")
			for app, appState := range askedStates {
				log.Info(helper.BuildPrefix(app), appState.StatusMessage())
			}
		}
	case "list":
		var result []string
		for app := range configuration.Settings.AppDefinitions {
			result = append(result, app)
		}
		log.Infoln("Available apps:", strings.Join(result, ","))
	case "install", "update", "upgrade":
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
				))
			if exitCode != EXIT_OK && !(*flagOptimist) {
				return exitCode
			}
		}
		return 0

	default:
		log.Errorln("Unknown action", action)
		return EXIT_UNKNOWN_ACTION
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
