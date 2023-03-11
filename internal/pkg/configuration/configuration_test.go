package configuration

import (
	"embed"
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"testing"
)

//go:embed configuration_test_embeddedDefs
var embeddedDefs embed.FS

func TestLoadWithGlobalCustomAndMeta(t *testing.T) {
	Settings = data.NewSettings()
	log.EnableLevelsByNumber(10)
	//log.Println(os.Getwd())

	//GIVEN
	testDataPath := "../../../test/data/"
	customPath := "configuration_test_customDefs"
	AppDefinitionDirectoryName = "configuration_test_embeddedDefs"

	//WHEN
	Load(testDataPath+"global_test.toml", customPath, embeddedDefs)

	//THEN
	assert.Len(t, Settings.AppDefinitions, 7)

	//Global Settings
	assert.ContainsKey(t, Settings.AppDefinitions, "ccleaner")
	app := Settings.AppDefinitions["ccleaner"]
	assert.Equal(t, "606", app.Version)
	assert.Equal(t, Settings.GithubApiKey, "12345")

	//TODO custom defs
	assert.ContainsKey(t, Settings.AppDefinitions, "customJson")
	assert.ContainsKey(t, Settings.AppDefinitions, "customToml")

	//App Defs from META
	assert.ContainsKey(t, Settings.AppDefinitions, "metaJson1")
	assert.ContainsKey(t, Settings.AppDefinitions, "metaJson2")
	assert.ContainsKey(t, Settings.AppDefinitions, "metaToml")

}

func TestLoadWithMetaOnly(t *testing.T) {
	Settings = data.NewSettings()
	log.EnableLevelsByNumber(10)
	//log.Println(os.Getwd())

	//GIVEN
	AppDefinitionDirectoryName = "configuration_test_embeddedDefs"

	//WHEN
	Load("404", "404", embeddedDefs)

	//THEN
	assert.Len(t, Settings.AppDefinitions, 3)

}
