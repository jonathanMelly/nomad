package configuration

import (
	"bufio"
	"github.com/gologme/log"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json"
	"github.com/gookit/config/v2/toml"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/internal/pkg/iohelper"
	"io"
	"os"
	"path"
	"strings"
)

var settings = data.NewSettings()

func Load(globalSettingsPath string, customDefinitionsDirectory string, binaryPath string) {
	loadGlobalSettings(func(config2 *config.Config) {
		config2.BindStruct("", &settings)
		if log.GetLevel("debug") {
			for k, _ := range settings.AppDefinitions {
				log.Debugln("Added", k, "definition")
			}
		}
	}, globalSettingsPath)

	ghKeyFromEnv := os.Getenv("GITHUB_PAT")
	if settings.GithubApiKey == "" && ghKeyFromEnv != "" {
		log.Debugln("Using github key from ENV")
		settings.GithubApiKey = ghKeyFromEnv
	} else if ghKeyFromEnv != "" && settings.GithubApiKey != ghKeyFromEnv {
		log.Debugln("Github key from config file and ENV differ... Using config file")
	}

	loadCustomAppDefinitions(customDefinitionsDirectory)

	//Embedded defs
	loadEmbeddedAppDefinitions(binaryPath)
}

func loadEmbeddedAppDefinitions(binaryPath string) {
	log.Traceln("Looking into ", binaryPath, "for app definition metadata")
	exeFile, err := os.Open(binaryPath)
	if err != nil {
		log.Errorln("Cannot read binary", binaryPath, "|", err)
		return
	}
	defer exeFile.Close()
	exeStat, err := exeFile.Stat()
	if err != nil {
		log.Errorln("Cannot stat binary", binaryPath, "|", err)
		return
	}

	reader := bufio.NewReader(exeFile)
	if exeStat.Size() > 1024*100 { //Test files do not contain binary...
		reader.Discard(int(float32(exeStat.Size()) * 0.8)) //config data should not be > than 80% of binary iohelper...
	}
	buf := make([]byte, 16)
	var exeContent = strings.Builder{}
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Error(err)
				return
			}
			break
		}
		exeContent.WriteString(string(buf[0:n]))
	}

	split := strings.Split(exeContent.String(), "#NOMAD#")
	if len(split) != 2 {
		log.Debugln("No metadata in binary", binaryPath)
	} else {
		nomadData := split[1]
		split = strings.Split(nomadData, "#NOMAD_TOML#")
		if len(split) != 2 {
			log.Errorln("Bad metadata in binary", binaryPath)
		} else {
			jsonContent := split[0]
			importFromConfigString(config.JSON, jsonContent, func(jsonConfig *config.Config) {
				jsonApps := data.JsonApps{}
				err := jsonConfig.BindStruct("", &jsonApps)
				if err != nil {
					log.Errorln("Cannot bind JsonApps struct from config", "|", err)
				}
				for _, definition := range jsonApps.Definitions {
					fillDefinitions(definition.ApplicationName, definition)
				}
			})

			tomlContent := split[1]
			importFromConfigString(config.Toml, tomlContent, func(tomlConfig *config.Config) {
				addTomlAppDefinitionsFromConfig(tomlConfig)
			})
		}
	}
}

func loadCustomAppDefinitions(customDefinitionsDirectory string) {
	//Custom Definitions files (json or toml)
	if customDefinitionsDirectory != "" && iohelper.FileOrDirExists(customDefinitionsDirectory) {
		log.Debugln("Looking into", customDefinitionsDirectory, "for custom app definitions")
		files, err := os.ReadDir(customDefinitionsDirectory)
		if err != nil {
			log.Errorln("Cannot read", customDefinitionsDirectory, "|", err)
			return
		}

		tomlConfig := initConfig()
		jsonMerge := strings.Builder{} //merge json contents to reduce config.Load mechanism...

		for _, f := range files {
			if !f.IsDir() {
				appDefinitionPath := path.Join(customDefinitionsDirectory, f.Name())
				if strings.HasSuffix(f.Name(), ".json") {
					json, err := os.ReadFile(appDefinitionPath)
					if jsonMerge.String() == "" {
						jsonMerge.WriteString(`{"apps":[`)
					} else {
						jsonMerge.WriteString(`,`)
					}
					jsonMerge.Write(json)
					if err != nil {
						log.Errorln("Cannot load custom json config", appDefinitionPath, "|", err)
					}
				} else if strings.HasSuffix(f.Name(), ".toml") {
					err := tomlConfig.LoadFiles(appDefinitionPath)
					if err != nil {
						log.Errorln("Cannot load custom toml config", appDefinitionPath, "|", err)
					}
				}
			}
		}

		//JSON
		jsonMerge.WriteString(`]}`)
		importFromConfigString(config.JSON, jsonMerge.String(), func(jsonConfig *config.Config) {
			addJsonAppDefinitionsFromConfig(jsonConfig)
		})

		//TOML
		addTomlAppDefinitionsFromConfig(tomlConfig)

	} else {
		log.Debugln("No custom app definition found in", customDefinitionsDirectory)
	}
}

func addTomlAppDefinitionsFromConfig(tomlConfig *config.Config) {
	tomlApps := data.Apps{}
	err := tomlConfig.BindStruct("", &tomlApps)
	if err != nil {
		log.Errorln("Cannot bind Apps struct from config", "|", err)
	}
	//Cannot use copy as we don’t want to override (first config come, first served)
	for app, definition := range tomlApps.Definitions {
		fillDefinitions(app, definition)
	}
}

func addJsonAppDefinitionsFromConfig(jsonConfig *config.Config) {
	jsonApps := data.JsonApps{}
	err := jsonConfig.BindStruct("", &jsonApps)
	if err != nil {
		log.Errorln("Cannot bind JsonApps struct from config", "|", err)
	}
	for _, definition := range jsonApps.Definitions {
		fillDefinitions(definition.ApplicationName, definition)
	}
}

func fillDefinitions(app string, definition data.AppDefinition) {
	_, exist := settings.AppDefinitions[app]
	if !exist {
		log.Traceln("Adding", app, "definition")

		//For TOML, appname is in key...
		if definition.ApplicationName == "" {
			definition.ApplicationName = app
		}
		settings.AppDefinitions[app] = definition
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

func loadGlobalSettings(do func(config2 *config.Config), paths ...string) {
	log.Traceln("Loading global config from", paths)
	configTmp := initConfig()
	for _, path := range paths {
		if iohelper.FileOrDirExists(path) {
			err := configTmp.LoadFiles(path)
			if err != nil {
				log.Errorln("Cannot read config iohelper", path, "|", err)
			} else {
				do(configTmp)
				configTmp.ClearAll()
			}
		} else {
			log.Debugln(path, "not found/existing, skipping")
		}
	}
}

func importFromConfigString(format string, content string, do func(config2 *config.Config)) {
	configTmp := initConfig()
	err := configTmp.LoadStrings(format, content)
	if err != nil {
		log.Errorln("Cannot read", format, "config content", "|", err)
	} else {
		do(configTmp)
		configTmp.ClearAll()
	}
}
