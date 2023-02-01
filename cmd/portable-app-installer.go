// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"flag"
	"fmt"
	"github.com/jonathanMelly/portable-app-installer/installer"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func customUsage() {
	appName := filepath.Base(os.Args[0])
	fmt.Printf("Usage: %s install|update|status [OPTIONS] appName\n\nOPTIONS:\n", appName)
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

	flagEnvVarForAppsLocation := flag.String("envvar", "%apps%", "If not empty, will be used for shortcuts generation as base path... (to allow easy switch)")

	flagForceExtract := flag.Bool("force", false, "Remove any similar existing version")
	flagSkipDownload := flag.Bool("skip", true, "Skip download if corresponding archive is already present")
	flagLatestVersion := flag.Bool("latest", true, "If version URL is set, check and use latest version available")

	flagOptimist := flag.Bool("optimist", true, "If true and multiple config given, continue after one failed")

	flagConfirm := flag.Bool("confirm", true, "Asks user to confirm operation")

	flagArchivesSubDir := flag.String("archives", "archives", "Set archives subdir")

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(52)
	}

	action := flag.Arg(0)

	switch action {
	case "install":
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
							installer.Run(path.Join(*flagConfig, f.Name()), *flagVersion,
								*flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation, *flagArchivesSubDir,
								*flagLatestVersion, *flagConfirm))
						log.Println("")
						if exitCode > 0 && !(*flagOptimist) {
							os.Exit(exitCode)
						}
					}
				}
			} else {
				log.Println("/!\\ Error : no app name given")
				flag.Usage()
				os.Exit(51)
			}

		} else {
			// Run the automation
			exitCode := HandleRun(installer.Run(configFile, *flagVersion,
				*flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation,
				*flagArchivesSubDir, *flagLatestVersion, *flagConfirm))
			os.Exit(exitCode)
		}
	case "update":
	default:
		flag.Usage()
		os.Exit(53)
	}

}

func HandleRun(err error, errorMessage string, exitCode int) int {
	if exitCode == 0 {
		log.Println("SUCCESS !")
	} else {
		if errorMessage != "" {
			log.Print(errorMessage)
			if err != nil {
				log.Println(" |")
			}
		} else if err != nil {
			fmt.Errorf("%w", err)
		}

	}

	return exitCode
}
