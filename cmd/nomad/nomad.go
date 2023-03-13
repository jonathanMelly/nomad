// Initial Author
// Copyright 2015 Joseph Spurrier
// Author: Joseph Spurrier (http://josephspurrier.com)
// License: http://www.apache.org/licenses/LICENSE-2.0.html

// New version J. Melly

package main

import (
	"embed"
	_ "embed"

	"github.com/jonathanMelly/nomad/internal/app/cli"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"os"
)

// x-release-please-start-version
var version = "1.2.1"

//x-release-please-end

//go:embed app-definitions
var embeddedDefs embed.FS

//go:embed ghkey.txt
var githubApiKey string

func main() {
	cli.Version = version
	cli.EmbeddedDefinitions = embeddedDefs
	//Sets default key (can be later overridden)
	configuration.Settings.GithubApiKey = githubApiKey
	os.Exit(cli.Main())
}
