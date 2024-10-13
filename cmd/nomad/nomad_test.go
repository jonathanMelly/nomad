package main

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	version2 "github.com/jonathanMelly/nomad/pkg/version"
	"os"
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
			wg.Add(1)
			go checkDownloadableAsset(t, def)
		}
	}
	wg.Wait()

}

func checkDownloadableAsset(t *testing.T, def *data.AppDefinition) {

	defVersion, _ := version2.FromString(def.Version)
	downloadURL := defVersion.FillVersionsPlaceholders(def.DownloadUrl)

	client, err := helper.BuildAndDoHttp(downloadURL, "HEAD", def.SslIgnoreBadCert)

	if assert.NoError(t, err, "http client error url "+downloadURL) &&
		assert.NotNil(t, client, "app", def.ApplicationName, " failed : ", "http client for url "+downloadURL+" should not be nil") {

		expectedCode := 200
		if IsBogus(def) {
			expectedCode = 403
		}

		assert.Equal(t, expectedCode, client.StatusCode, downloadURL+" should return a 200 status code upon HEAD request")
	}
	wg.Done()
}

func IsBogus(def *data.AppDefinition) bool {
	return def.ApplicationName == "resourcehacker" && os.Getenv("GITHUB_ACTIONS") == "true"
}
