package configuration

import (
	"embed"
	"fmt"
	"github.com/gologme/log"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/helper"
	"github.com/jonathanMelly/nomad/pkg/version"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var Settings = data.NewSettings()
var AppDefinitionDirectoryName = "app-definitions"

const DefaultShortcutsDir = "shortcuts"

var AppPath = "apps"

func Load(globalSettingsPath string, customDefinitionsDirectory string, embeddedSrc embed.FS) {

	//Load key from ENV
	ghKeyFromEnv := os.Getenv("GITHUB_PAT")
	if ghKeyFromEnv != "" && ghKeyFromEnv != Settings.GithubApiKey {
		log.Debugln("Using github key from ENV")
		Settings.GithubApiKey = ghKeyFromEnv
	}

	//Key can be overridden here
	loadGlobalSettings(func(config2 *config.Config) {
		err := config2.BindStruct("", &Settings)
		if err != nil {
			log.Errorln("Cannot bind struct", err)
		} else if log.GetLevel("debug") {
			for k := range Settings.AppDefinitions {
				log.Debugln("Added", k, "custom definition from", globalSettingsPath)
			}
		}
	}, globalSettingsPath)
	if log.GetLevel("debug") && ghKeyFromEnv != Settings.GithubApiKey {
		log.Debugln("Overriding github key with config")
	}

	loadCustomAppDefinitions(customDefinitionsDirectory)

	if &embeddedSrc != nil {
		log.Traceln("Loading embedded definitions")
		LoadEmbeddedDefinitions(embeddedSrc)
	}

}

func LoadEmbeddedDefinitions(embeddedSrc embed.FS) {
	loadAppDefinitions("embedded", AppDefinitionDirectoryName, embeddedSrc)
}

func loadCustomAppDefinitions(customDefinitionsDirectory string) {
	//Custom Definitions files (json or toml)
	if customDefinitionsDirectory != "" && helper.FileOrDirExists(customDefinitionsDirectory) {
		log.Debugln("Looking into", customDefinitionsDirectory, "for custom app definitions")
		wd, err := os.Getwd()
		if err != nil {
			log.Errorln("Cannot get current dir", err)
			return
		}
		loadAppDefinitions("custom", customDefinitionsDirectory, os.DirFS(wd))

	} else {
		log.Debugln("No custom app definition found in", customDefinitionsDirectory)
	}
}

func loadAppDefinitions(sourceIdentifier string, directoryPath string, fs2 fs.FS) {

	files, err := fs.ReadDir(fs2, directoryPath)
	if err != nil {
		log.Errorln("Cannot read dir", directoryPath, "with fs", fs2, "|", err)
		return
	}

	jsonMerge := strings.Builder{} //merge json contents to reduce config.Load mechanism...
	tomlMerge := strings.Builder{}
	for _, f := range files {
		if !f.IsDir() {
			appDefinitionPath := path.Join(directoryPath, f.Name())
			appNameFromFilename := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			const APP_NAME_ATTRIBUTE = "ApplicationName"
			if strings.HasSuffix(f.Name(), ".json") {
				jsonContent, err := fs.ReadFile(fs2, appDefinitionPath)
				if err != nil {
					log.Errorln("Cannot load json", sourceIdentifier, "config", appDefinitionPath, "|", err)
					continue
				}
				if jsonMerge.String() == "" {
					jsonMerge.WriteString(`{"apps":[`)
				} else {
					jsonMerge.WriteString(`,`)
				}

				jsonString := string(jsonContent)
				if !strings.Contains(jsonString, APP_NAME_ATTRIBUTE) {
					log.Traceln("Adding", APP_NAME_ATTRIBUTE, "from filename", appNameFromFilename)
					jsonContent = []byte(strings.Replace(jsonString, "{",
						fmt.Sprintf(`{\n\t"%s"":"%s,\n"`, APP_NAME_ATTRIBUTE, appNameFromFilename), 1))
				}
				jsonMerge.Write(jsonContent)

			} else if strings.HasSuffix(f.Name(), ".toml") {
				tomlContent, err := fs.ReadFile(fs2, appDefinitionPath)
				if err != nil {
					log.Errorln("Cannot load toml", sourceIdentifier, "config", appDefinitionPath, "|", err)
					continue
				}

				if !strings.HasPrefix(string(tomlContent), "[apps.") {
					log.Traceln("Adding", APP_NAME_ATTRIBUTE, "from filename", appNameFromFilename)
					tomlMerge.WriteString(fmt.Sprintf("[apps.%s]\n", appNameFromFilename))
				}
				tomlMerge.Write(tomlContent)
				tomlMerge.WriteString(fmt.Sprintln())
			}
		}
	}

	//JSON
	jsonMerge.WriteString(`]}`)
	importFromConfigString(config.JSON, jsonMerge.String(), func(jsonConfig *config.Config) {
		addJsonAppDefinitionsFromConfig(sourceIdentifier, jsonConfig)
	})

	//TOML
	importFromConfigString(config.Toml, tomlMerge.String(), func(tomlConfig *config.Config) {
		addTomlAppDefinitionsFromConfig(sourceIdentifier, tomlConfig)
	})

}

func addTomlAppDefinitionsFromConfig(identifier string, tomlConfig *config.Config) {
	tomlApps := data.Apps{}
	err := tomlConfig.BindStruct("", &tomlApps)
	if err != nil {
		log.Errorln("Cannot bind Apps struct from config", "|", err)
	}
	//Cannot use copy as we don’t want to override (first config come, first served)
	for app, definition := range tomlApps.Definitions {
		fillDefinitions(identifier, app, definition)
	}
}

func addJsonAppDefinitionsFromConfig(identifier string, jsonConfig *config.Config) {
	jsonApps := data.JsonApps{}
	err := jsonConfig.BindStruct("", &jsonApps)
	if err != nil {
		log.Errorln("Cannot bind JsonApps struct from config", "|", err)
	}
	for _, definition := range jsonApps.Definitions {
		fillDefinitions(identifier, definition.ApplicationName, definition)
	}
}

func fillDefinitions(identifier string, app string, definition data.AppDefinition) {
	//Old config format, should be removed end of year 2023
	if strings.Contains(app, version.VERSION_PLACEHOLDER) {
		app = app[0:strings.LastIndex(app, "-")]
		definition.ApplicationName = app
		log.Warnln("Please upgrade config : remove -{{VERSION}} from app name")
	}

	_, exist := Settings.AppDefinitions[app]
	if !exist {
		log.Traceln("Adding", app, "definition from", identifier)

		//For TOML, appname is in key...
		if definition.ApplicationName == "" {
			definition.ApplicationName = app
		}

		if valid, err := definition.IsValid(); !valid {
			log.Warnln("Invalid app definition", app, "|", err, "->discarding")
		} else {
			Settings.AppDefinitions[app] = &definition
		}

	} else {
		log.Traceln(app, "already defined->not adding it")
	}

}

func initConfig() *config.Config {
	ephemeralConfig := config.New("apps")
	ephemeralConfig.WithOptions(config.ParseEnv)
	ephemeralConfig.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig.TagName = "json"
	})
	ephemeralConfig.AddDriver(toml.Driver)
	ephemeralConfig.AddDriver(json.Driver)

	return ephemeralConfig
}

func loadGlobalSettings(do func(config2 *config.Config), settingsPaths ...string) {
	log.Traceln("Loading global config from", settingsPaths)
	configTmp := initConfig()
	for _, settingsPath := range settingsPaths {
		if helper.FileOrDirExists(settingsPath) {
			err := configTmp.LoadFiles(settingsPath)
			if err != nil {
				log.Errorln("Cannot read config helper", settingsPath, "|", err)
			} else {
				do(configTmp)
				configTmp.ClearAll()
			}
		} else {
			log.Debugln(settingsPath, "not found/existing, skipping")
		}
	}
}

func importFromConfigString(format string, content string, do func(config2 *config.Config)) {
	configTmp := initConfig()
	err := configTmp.LoadStrings(format, content)
	if err != nil {
		log.Errorln("Cannot read", format, "config content", content, "|", err)
	} else {
		do(configTmp)
		configTmp.ClearAll()
	}
}
