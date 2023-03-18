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
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"os"
)

// x-release-please-start-version
var version = "1.5.2"

//x-release-please-end

//go:generate go run ./generate/

//go:embed date.txt
var buildDate string

//go:embed app-definitions
var embeddedDefs embed.FS

//go:embed ghkey.txt
var githubApiKey string

func main() {
	cli.Version = fmt.Sprint(version, " [", buildDate, "]", " (", commit, ")")
	cli.EmbeddedDefinitions = embeddedDefs
	//Sets default key (can be later overridden)
	configuration.Settings.GithubApiKey = githubApiKey
	os.Exit(cli.Main())
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
