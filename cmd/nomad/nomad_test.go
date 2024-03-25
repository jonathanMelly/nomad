package main

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	version2 "github.com/jonathanMelly/nomad/pkg/version"
	"strings"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestValidateDefaultAppDefinitions(t *testing.T) {
	log.EnableLevelsByNumber(10)
	//Check load
	configuration.LoadEmbeddedDefinitions(embeddedDefs)
	definitions := configuration.Settings.AppDefinitions
	assert.Gt(t, len(definitions), 20)

	//Verify with filesystem
	foundFiles, err := embeddedDefs.ReadDir(configuration.AppDefinitionDirectoryName)
	assert.NoError(t, err)
	assert.Equal(t, len(foundFiles), len(definitions)+1 /*add current dir*/)

	//Validate confg
	for _, def := range definitions {
		valid, err := def.IsValid()
		assert.True(t, valid)
		assert.NoError(t, err)

		//check that url exists
		if strings.HasPrefix(def.DownloadUrl, "http") {
			checkDownloadableAsset(t, def)
		}
	}

}

func checkDownloadableAsset(t *testing.T, def *data.AppDefinition) {
	log.Debug("checking  url for ", def.ApplicationName)

	defVersion, _ := version2.FromString(def.Version)
	downloadURL := defVersion.FillVersionsPlaceholders(def.DownloadUrl)

	client, err := helper.BuildAndDoHttp(downloadURL, "HEAD", def.SslIgnoreBadCert)
	assert.NoError(t, err)
	assert.NotNil(t, client, "http client result for url"+downloadURL+" should not be nil")
	assert.Equal(t, 200, client.StatusCode, downloadURL+" should return a 200 status code upon HEAD request")

	log.Debugln(" --- OK")
}
