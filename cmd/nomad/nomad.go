// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"embed"
	_ "embed"
	"fmt"
	"runtime/debug"

	"github.com/jonathanMelly/nomad/internal/app/cli"
	"os"
)

// x-release-please-start-version
var version = "1.12.0"

//x-release-please-end

//go:generate go run ./generate/

//go:embed date.txt
var buildDate string

//go:embed app-definitions
var embeddedDefs embed.FS

//go:embed ghkey.txt
var githubApiKey string

func main() {
	os.Exit(cli.Main(
		embeddedDefs,
		githubApiKey,
		version,
		fmt.Sprint(" [", buildDate, "]", " (", commit, ")"),
	))
}

var commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" /*vcs.time also available*/ {
				return setting.Value
			}
		}
	}
	return ""
}()
