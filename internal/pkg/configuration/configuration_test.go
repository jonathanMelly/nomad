package configuration

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"testing"
)

func TestLoadWithGlobalCustomAndMeta(t *testing.T) {
	settings = data.NewSettings()
	log.EnableLevelsByNumber(10)
	//log.Println(os.Getwd())

	//GIVEN
	testDataPath := "../../../test/data/"
	customPath := testDataPath + "customDefs"

	//WHEN
	Load(testDataPath+"global_test.toml", customPath, testDataPath+"fakePatchedBinaryWithMeta")

	//THEN
	assert.Len(t, settings.AppDefinitions, 7)

	//Global settings
	assert.ContainsKey(t, settings.AppDefinitions, "ccleaner")
	app := settings.AppDefinitions["ccleaner"]
	assert.Equal(t, "606", app.Version)
	assert.Equal(t, settings.GithubApiKey, "12345")

	//TODO custom defs
	assert.ContainsKey(t, settings.AppDefinitions, "customJson")
	assert.ContainsKey(t, settings.AppDefinitions, "customToml")

	//App Defs from META
	assert.ContainsKey(t, settings.AppDefinitions, "metaJson1")
	assert.ContainsKey(t, settings.AppDefinitions, "metaJson2")
	assert.ContainsKey(t, settings.AppDefinitions, "metaToml")

}

func TestLoadWithMetaOnly(t *testing.T) {
	settings = data.NewSettings()
	log.EnableLevelsByNumber(10)
	//log.Println(os.Getwd())

	//GIVEN
	testDataPath := "../../../test/data/"

	//WHEN
	Load("404", "404", testDataPath+"fakePatchedBinaryWithMeta")

	//THEN
	assert.Len(t, settings.AppDefinitions, 3)

}
