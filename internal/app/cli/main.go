package cli

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/installer"
	"github.com/jonathanMelly/nomad/internal/pkg/state"
	"math"
	"os"
	"path/filepath"
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
}

func printVersion() {
	_, err := fmt.Println(exeName, Version)
	if err != nil {
		log.Errorln(err)
	}
}

func Main() int {

	flag.Usage = customUsage
	// Overwrite version
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions iohelper")
	flagDefinitionsDirectory := flag.String("definitions", "app-definitions", "Specify directory for custom app definitions files (relative to current dir and not outside...)")

	flagEnvVarForAppsLocation := flag.String("shortcutLocation", "", "If not empty, will be used for shortcuts generation as base path... (to allow easy switch)")

	flagForceExtract := flag.Bool("force", false, "Remove any similar existing version")
	flagSkipDownload := flag.Bool("skip", true, "Skip download if corresponding archive is already present")
	flagLatestVersion := flag.Bool("latest", true, "If version URL is set, check and use latest version available")

	flagOptimist := flag.Bool("optimist", true, "If true and multiple config given, continue after one failed")

	flagConfirm := flag.Bool("confirm", true, "Asks user to confirm operation")

	flagArchivesSubDir := flag.String("archives", "archives", "Set archives subdir")

	flagVerbose := flag.Bool("verbose", false, "Gives info for debug")

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(52)
	}

	/*
	   Level 10 = panic, fatal, error, warn, info, debug, & trace
	   Level 5 = panic, fatal, error, warn, info, & debug
	   Level 4 = panic, fatal, error, warn, & info
	   Level 3 = panic, fatal, error, & warn
	   Level 2 = panic, fatal & error
	   Level 1 = panic, fatal
	*/
	if *flagVerbose {
		log.EnableLevelsByNumber(10)
	} else {
		log.EnableLevelsByNumber(4)
	}
	configuration.Settings.ArchivesDirectory = *flagArchivesSubDir

	//Prefix is used to show app name...
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	log.Traceln("Args", flag.Args())

	configuration.Load("nomad.toml", *flagDefinitionsDirectory, EmbeddedDefinitions)

	action := flag.Arg(0)

	switch action {

	case "version":
		printVersion()
		key := configuration.Settings.GithubApiKey
		log.Debug("Using token ", key[0:int(math.Min(float64(len(key)), 15))], "...\n")
		return 0

	case "status", "install", "update":
		apps := flag.Args()[1:]
		if len(apps) == 0 || apps[0] == "all" {
			log.Traceln("No app given, searching for all apps in", installer.AppPath)
			apps = []string{}

			for k := range state.GetCurrentApps(installer.AppPath) {
				apps = append(apps, k)
			}
		}
		log.Debugln("Selected apps:", apps)
		for _, app := range apps {
			definition, exist := configuration.Settings.AppDefinitions[app]
			if exist {
				log.Debugln("Processing", app)
				exitCode := HandleRun(
					installer.Run(action, definition, *flagVersion,
						*flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation, *flagArchivesSubDir,
						*flagLatestVersion, *flagConfirm))
				if exitCode > 0 && !(*flagOptimist) {
					return exitCode
				}
			} else {
				log.Warnln("Unknown app", app, "->skipping")
			}

		}
		return 0

	default:
		log.Errorln("Unknown action", action)
		return 67

	}

	return 0

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
