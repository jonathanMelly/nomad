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
	"github.com/jonathanMelly/nomad/installer"
	"io/ioutil"
	"os"
	"path"
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
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions file")
	flagConfig := flag.String("configs", "", "Runs all json in given folder")

	flagEnvVarForAppsLocation := flag.String("shortcutLocation", "", "If not empty, will be used for shortcuts generation as base path... (to allow easy switch)")

	flagForceExtract := flag.Bool("force", false, "Remove any similar existing version")
	flagSkipDownload := flag.Bool("skip", true, "Skip download if corresponding archive is already present")
	flagLatestVersion := flag.Bool("latest", true, "If version URL is set, check and use latest version available")

	flagOptimist := flag.Bool("optimist", true, "If true and multiple config given, continue after one failed")

	flagConfirm := flag.Bool("confirm", true, "Asks user to confirm operation")

	flagArchivesSubDir := flag.String("archives", "archives", "Set archives subdir")

	/*
	   Level 10 = panic, fatal, error, warn, info, debug, & trace
	   Level 5 = panic, fatal, error, warn, info, & debug
	   Level 4 = panic, fatal, error, warn, & info
	   Level 3 = panic, fatal, error, & warn
	   Level 2 = panic, fatal & error
	   Level 1 = panic, fatal
	*/
	flagVerbose := flag.Bool("verbose", false, "Gives info for debug")

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(52)
	}

	if *flagVerbose {
		log.EnableLevelsByNumber(10)
	} else {
		log.EnableLevelsByNumber(4)
	}
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	action := flag.Arg(0)

	configFile := flag.Arg(1)

	if configFile == "" {
		if *flagConfig != "" {
			files, err := ioutil.ReadDir(*flagConfig)
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					exitCode := HandleRun(
						installer.Run(action, path.Join(*flagConfig, f.Name()), *flagVersion,
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
		exitCode := HandleRun(installer.Run(action, configFile, *flagVersion,
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
