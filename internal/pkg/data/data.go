package data

import (
	"errors"
	"fmt"
	"github.com/jonathanMelly/nomad/pkg/version"
	"regexp"
	"strings"
)

type AppStates struct {
	States map[string]*AppState
}

func NewAppStates() *AppStates {
	return &AppStates{
		States: map[string]*AppState{},
	}
}

type AppState struct {
	Definition     AppDefinition
	SymlinkFound   bool
	CurrentVersion *version.Version
	TargetVersion  *version.Version
}

type Settings struct {
	MyApps            []string                  `json:"myapps"`
	GithubApiKey      string                    `json:"githubApiKey"`
	AppDefinitions    map[string]*AppDefinition `json:"apps"`
	ArchivesDirectory string                    `json:"archivesDirectory"`
}

func NewSettings() *Settings {
	return &Settings{
		MyApps:         []string{},
		GithubApiKey:   "",
		AppDefinitions: map[string]*AppDefinition{},
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
	Symlink           string            `json:"Symlink"` //use it instead of appname for symlink (if given)
	Shortcut          string            `json:"Shortcut"`
	ShortcutIcon      string            `json:"ShortcutIcon"`
	DownloadUrl       string            `json:"DownloadUrl"`
	ExtractRegExList  []string          `json:"ExtractRegExList"`
	CreateFolders     []string          `json:"CreateFolders"`
	CreateFiles       map[string]string `json:"CreateFiles"`
	MoveObjects       map[string]string `json:"MoveObjects"`
	RestoreFiles      []string          `json:"RestoreFiles"` //Copy/Paste (overwrite) files from previous symlinked directory (needs symlink)
	Validated         bool
	ExtractRegex      *regexp.Regexp
}

func (definition *AppDefinition) Validate() error {
	var errs []string
	if definition.ApplicationName == "" {
		errs = append(errs, "missing application name")
	}
	if strings.Contains(definition.ApplicationName, "-") {
		errs = append(errs, "app name cannot contain - (dash), please replace with something else (ex. _)")
	}
	if definition.DownloadExtension == "" {
		errs = append(errs, "missimg download extension")
	}
	if definition.VersionCheck.Url == "" && definition.Version == "" {
		errs = append(errs, "missing version info (either fixed or by url)")
	}

	regex, err := combineRegex(definition.ExtractRegExList)
	if err != nil {
		errs = append(errs, fmt.Sprint("invalid regex for zip files ", definition.ExtractRegExList, " | ", err))
	} else {
		definition.ExtractRegex = regex
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprint("data errors: ", strings.Join(errs, ",")))
	} else {
		definition.Validated = true
	}

	return nil
}

type VersionCheck struct {
	Url              string `json:"Url"`
	RegEx            string `json:"RegEx"`
	UseLatestVersion bool   `json:"UseLatestVersion"`
}

// CombineRegex will take a string array of regular expressions and compile them
// into a single regular expressions
func combineRegex(s []string) (*regexp.Regexp, error) {
	joined := strings.Join(s, "|")

	re, err := regexp.Compile(joined)
	if err != nil {
		return nil, err
	}

	return re, nil
}
