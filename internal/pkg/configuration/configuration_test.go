package configuration

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"testing"
)

func TestLoadWithGlobalCustomAndMeta(t *testing.T) {
	Settings = data.NewSettings()
	log.EnableLevelsByNumber(10)
	//log.Println(os.Getwd())

	//GIVEN
	testDataPath := "../../../test/data/"
	customPath := testDataPath + "customDefs"

	//WHEN
	Load(testDataPath+"global_test.toml", customPath, testDataPath+"fakePatchedBinaryWithMeta")

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
	testDataPath := "../../../test/data/"

	//WHEN
	Load("404", "404", testDataPath+"fakePatchedBinaryWithMeta")

	//THEN
	assert.Len(t, Settings.AppDefinitions, 3)

}
