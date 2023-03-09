// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"flag"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/installer"
	"os"
	"path/filepath"
	"strings"
)

func customUsage() {
	appName := filepath.Base(os.Args[0])
	fmt.Printf("Usage: %s install|status [OPTIONS] appName\n\nOPTIONS:\n", appName)
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("\t", appName, "install rclone")
	fmt.Println("\t", appName, "-confirm=false install filezilla")
}

func main() {

	flag.Usage = customUsage
	// Overwrite version
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions iohelper")
	flagDefinitionsDirectory := flag.String("definitions", "app-definitions", "Specify directory for custom app definitions files")

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

	//Prefix is used to show app name...
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	exe, err := os.Executable()
	if err != nil {
		log.Panic("Cannot retrieve current process", err)
	}

	configuration.Load(*flagDefinitionsDirectory, "nomad.toml", exe)

	action := flag.Arg(0)
	configFile := flag.Arg(1)

	if configFile == "" {
		if *flagDefinitionsDirectory != "" {
			files, err := os.ReadDir(*flagDefinitionsDirectory)
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					//conf := path.Join(*flagDefinitionsDirectory, f.Name())
					exitCode := HandleRun(
						installer.Run(action /*TODO*/, data.AppDefinition{}, *flagVersion,
							*flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation, *flagArchivesSubDir,
							*flagLatestVersion, *flagConfirm))
					//log.Println("")
					if exitCode > 0 && !(*flagOptimist) {
						os.Exit(exitCode)
					}
				}
			}
		} else {
			log.Error("/!\\ Error : no app name given")
			flag.Usage()
			os.Exit(51)
		}

	} else {
		// Run the automation
		exitCode := HandleRun(installer.Run(action /*TODO*/, data.AppDefinition{}, *flagVersion,
			*flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation,
			*flagArchivesSubDir, *flagLatestVersion, *flagConfirm))
		os.Exit(exitCode)
	}

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
