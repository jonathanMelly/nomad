package installer

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

// AppInfo contains the settings for the portable application
type AppInfo struct {
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

type VersionCheck struct {
	Url              string `json:"Url"`
	RegEx            string `json:"RegEx"`
	UseLatestVersion bool   `json:"UseLatestVersion"`
}

// ParseJSON parses the given bytes
func (appInfo *AppInfo) parseJSON(jsonBytes []byte) error {
	return json.Unmarshal([]byte(jsonBytes), &appInfo)
}

// LoadConfig returns the struct from the app-definitions file
func LoadConfig(configFile string) (*AppInfo, error) {
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

	// Create a new container
	pi := &AppInfo{}

	// Parse the app-definitions
	if err := pi.parseJSON(jsonBytes); err != nil {
		return nil, err
	}

	return pi, nil
}
