package data

import (
	"errors"
	"fmt"
)

type Settings struct {
	MyApps         []string                 `json:"myapps"`
	GithubApiKey   string                   `json:"githubApiKey"`
	AppDefinitions map[string]AppDefinition `json:"apps"`
}

func NewSettings() *Settings {
	return &Settings{
		MyApps:         []string{},
		GithubApiKey:   "",
		AppDefinitions: map[string]AppDefinition{},
	}
}

// JsonApps Used for embedded configs (added to binary during build)
type JsonApps struct {
	Definitions []AppDefinition `json:"apps"`
}

// Apps Directly used for TOML binding AND generally in the app (json is converted... as a map is more convenient)
type Apps struct {
	Definitions map[string]AppDefinition `json:"apps"`
}

// AppDefinition contains the settings for the portable application
type AppDefinition struct {
	ApplicationName   string            `json:"ApplicationName"`
	DownloadExtension string            `json:"DownloadExtension"`
	Version           string            `json:"Version"`
	VersionCheck      VersionCheck      `json:"VersionCheck"`
	RemoveRootFolder  bool              `json:"RemoveRootFolder"`
	RootFolderName    string            `json:"RootFolderName"`
	Symlink           string            `json:"Symlink"` //if specified add/update symlink to point to appname folder
	Shortcut          string            `json:"Shortcut"`
	ShortcutIcon      string            `json:"ShortcutIcon"`
	DownloadUrl       string            `json:"DownloadUrl"`
	ExtractRegExList  []string          `json:"ExtractRegExList"`
	CreateFolders     []string          `json:"CreateFolders"`
	CreateFiles       map[string]string `json:"CreateFiles"`
	MoveObjects       map[string]string `json:"MoveObjects"`
	RestoreFiles      []string          `json:"RestoreFiles"` //Copy/Paste (overwrite) files from previous symlinked directory (needs symlink)
}

func (definition *AppDefinition) Validate() error {
	var missing []string
	if definition.ApplicationName == "" {
		missing = append(missing, "application name")
	}
	if definition.DownloadExtension == "" {
		missing = append(missing, "download extension")
	}
	if len(missing) > 0 {
		return errors.New(fmt.Sprint("mandatory data missing: ", missing))
	}
	return nil
}

type VersionCheck struct {
	Url              string `json:"Url"`
	RegEx            string `json:"RegEx"`
	UseLatestVersion bool   `json:"UseLatestVersion"`
}

/*
// ParseJSON parses the given bytes
func (appInfo *AppDefinition) parseJSON(jsonBytes []byte) error {
	return json.Unmarshal([]byte(jsonBytes), &appInfo)
}

// LoadConfig returns the struct from the app-definitions file
func LoadConfig(configFile string) (*AppDefinition, error) {
	var err error
	var input = io.ReadCloser(os.Stdin)
	if input, err = os.Open(configFile); err != nil {
		return nil, err
	}

	// Read the app-definitions file
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		return nil, err
	}

	// Create a upcoming container
	pi := &AppDefinition{}

	// Parse the app-definitions
	if err := pi.parseJSON(jsonBytes); err != nil {
		return nil, err
	}

	return pi, nil
}
*/
