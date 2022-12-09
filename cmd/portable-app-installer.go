// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"flag"
	"github.com/jonathanMelly/portable-app-installer/installer"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func main() {

	// Overwrite version
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions file")
	flagConfig := flag.String("configs", "", "Runs all json in given folder")

	flagEnvVarForAppsLocation := flag.String("envvar", "%apps%", "If not empty, will be used for shortcuts generation as base path... (to allow easy switch)")

	flagForceExtract := flag.Bool("force", false, "Remove any similar existing version")
	flagSkipDownload := flag.Bool("skip", true, "Skip download if corresponding archive is already present")

	flagArchivesSubDir := flag.String("archives", "archives", "Set archives subdir")

	flag.Parse()

	configFile := flag.Arg(0)

	if configFile == "" {
		if *flagConfig != "" {
			files, err := ioutil.ReadDir(*flagConfig)
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					installer.Run(path.Join(*flagConfig, f.Name()), *flagVersion, *flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation, *flagArchivesSubDir)
					log.Println("")
				}
			}
		} else {
			log.Println("JSON Config file must be passed")
			os.Exit(1)
		}

	} else {
		// Run the automation
		installer.Run(configFile, *flagVersion, *flagForceExtract, *flagSkipDownload, *flagEnvVarForAppsLocation, *flagArchivesSubDir)
	}

}
