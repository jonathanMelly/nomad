package main

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"testing"
)

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
	}
}
