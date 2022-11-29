// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"flag"
	"log"
	"os"

	"github.com/jonathanMelly/portable-app-installer/installer"
)

func main() {

	// Overwrite version
	flagVersion := flag.String("version", "", "Overwrites the version in the app-definitions file")

	flag.Parse()

	configFile := flag.Arg(0)

	if configFile == "" {
		log.Println("JSON Config file must be passed")
		os.Exit(1)
	}

	// Run the automation
	installer.Run(configFile, *flagVersion)
}
