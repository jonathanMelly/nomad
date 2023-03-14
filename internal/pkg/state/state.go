package state

import (
	"github.com/chenhg5/collection"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/configuration"
	"github.com/jonathanMelly/nomad/pkg/version"
	"os"
	"strings"
)

func GetCurrentApps(directory string, exclude ...string) map[string]*version.Version {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	if exclude == nil {
		exclude = []string{configuration.Settings.ArchivesDirectory}
	}
	exclusions := collection.Collect(exclude)

	installedApps := map[string]*version.Version{}
	for _, f := range files {
		if f.IsDir() {
			log.Traceln("Inspecting dir", f.Name())
			//Hope that app name do not contain a dash ;-)
			split := strings.Split(f.Name(), "-")
			guessedApp := split[0]
			_, alreadyFound := installedApps[guessedApp]

			v, err := version.FromString(split[1])
			if err != nil {
				log.Errorln("Cannot get version of", guessedApp)
			}

			if !exclusions.Contains(f.Name()) && !alreadyFound {
				_, knownApp := configuration.Settings.AppDefinitions[guessedApp]
				if knownApp {

					installedApps[guessedApp] = v
				} else {
					log.Warnln("Unknown app", guessedApp)
				}
			}
		}
	}

	return installedApps
}
