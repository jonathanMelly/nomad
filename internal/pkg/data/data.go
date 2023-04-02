package data

import (
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/gookit/goutil/maputil"
	"regexp"
	"strings"
)

const GITHUB_GRAPHQL_URL = "https://api.github.com/graphql"
const GITHUB_PREFIX = "github"
const GITHUB_BASE_URL = "https://github.com/"

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
	//MANDATORY FIELDS
	Version     string `json:"Version"`
	DownloadUrl string `json:"DownloadUrl"` //without /, auto add tag_name for repo based app (see wsl2-ssh-pageant.toml)

	//Optional fields
	RepositoryUrl string `json:"RepositoryUrl"` //for easy config with github repos

	ApplicationName   string `json:"ApplicationName"`   //extracted from filename if missing
	DownloadExtension string `json:"DownloadExtension"` //extracted from download url if missing

	VersionCheck VersionCheck `json:"VersionCheck"` //optional
	Symlink      string       `json:"Symlink"`      //use it instead of appname for symlink (if given)
	Shortcut     string       `json:"Shortcut"`     //Optional
	ShortcutIcon string       `json:"ShortcutIcon"` //Optional

	ExtractRegExList []string          `json:"ExtractRegExList"` //Optional
	CreateFolders    []string          `json:"CreateFolders"`    //Optio
	CreateFiles      map[string]string `json:"CreateFiles"`
	NoAddVersionFile bool              //to avoid VERSION-{{VERSION}}.nomad file adding
	MoveObjects      map[string]string `json:"MoveObjects"`
	RestoreFiles     []string          `json:"RestoreFiles"` //Copy/Paste (overwrite) files from previous symlinked directory (needs symlink)

	//Internal stuff
	validated    bool
	extractRegex *regexp.Regexp
}

func (definition *AppDefinition) validateAndSetDefaults() error {
	var errs []string
	//NAME
	if definition.ApplicationName == "" {
		errs = append(errs, "missing application name")
	}

	if definition.Version == "" {
		errs = append(errs, "missing base version")
	}

	//SYMLINK
	//Sets default symlink to app name
	if definition.Symlink == "" {
		definition.Symlink = definition.ApplicationName
	}

	//Repository facilitation
	definition.fillInfosFromRepository(errs)

	//DOWNLOAD EXT
	definition.ComputeDownloadExtension()

	//VERSION
	if definition.VersionCheck.Url == "" && definition.Version == "" {
		errs = append(errs, "missing version info (either fixed or by url)")
	}

	//REGEX
	if definition.ExtractRegExList == nil || len(definition.ExtractRegExList) == 0 {
		definition.ExtractRegExList = []string{"(.*)"}
	}
	regex, err := combineRegex(definition.ExtractRegExList)
	if err != nil {
		errs = append(errs, fmt.Sprint("invalid regex for archive files ", definition.ExtractRegExList, " | ", err))
	} else {
		definition.extractRegex = regex
	}

	//Version file for easy see in explorer
	if !definition.NoAddVersionFile {
		const VersionFile = "VERSION-{{VERSION}}.nomad"
		if definition.CreateFiles == nil {
			definition.CreateFiles = map[string]string{VersionFile: "{{VERSION}}"}
		} else if !maputil.HasKey(definition.CreateFiles, VersionFile) {
			definition.CreateFiles[VersionFile] = "{{VERSION}}"
		}
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprint("data errors: ", strings.Join(errs, ",")))
	} else {
		definition.validated = true
		return nil
	}

}

func (definition *AppDefinition) fillInfosFromRepository(errs []string) {
	if definition.RepositoryUrl != "" {
		repoProvider, repoInfos, found := strings.Cut(definition.RepositoryUrl, ":")
		if found {
			switch repoProvider {
			case GITHUB_PREFIX:
				log.Traceln("Computing github infos for", repoInfos)
				_, _, ok := strings.Cut(repoInfos, "/")
				if ok {
					if definition.VersionCheck.Url == "" {
						definition.VersionCheck.Url = fmt.Sprint(repoProvider, ":", repoInfos)
					}
					if definition.VersionCheck.RegEx == "" {
						definition.VersionCheck.RegEx = fmt.Sprintf(`"tagName":"[^\d]*{{VERSION}}"`)
					}

					if !strings.HasPrefix(definition.DownloadUrl, "http") && !strings.HasPrefix(definition.DownloadUrl, "manual") {
						definition.DownloadUrl = fmt.Sprint(GITHUB_BASE_URL, repoInfos, "/releases/download/", definition.DownloadUrl)
					}

				} else {
					errs = append(errs, fmt.Sprint("bad github repository info ", repoProvider, " (missing owner or repo,syntax is github:owner/repo)"))
				}
			default:
				errs = append(errs, fmt.Sprint("unsupported repository provider ", repoProvider))
			}

		} else {
			errs = append(errs, fmt.Sprint("missing repository provider in RepositoryUrl ", definition.RepositoryUrl))
		}

	}
}

func (definition *AppDefinition) GetExtractRegex() *regexp.Regexp {
	return definition.extractRegex
}

func (definition *AppDefinition) IsValid() (bool, error) {
	if !definition.validated {
		err := definition.validateAndSetDefaults()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (definition *AppDefinition) ComputeDownloadExtension() {
	if definition.DownloadExtension == "" {
		const defaultExt = ".zip"
		if definition.DownloadUrl != "" {
			if !strings.HasPrefix(definition.DownloadUrl, "manual") { //Let manual extension be determined later
				lastPoint := strings.LastIndex(definition.DownloadUrl, ".")
				if lastPoint >= 0 {
					cutExtension, _, _ := strings.Cut(definition.DownloadUrl[lastPoint:], "?") //removes URL get params after filename
					definition.DownloadExtension = cutExtension
				} else {
					definition.DownloadExtension = defaultExt
				}
			}

		} else {
			definition.DownloadExtension = defaultExt
		}
	}
}

type VersionCheck struct {
	Url              string `json:"Url"`
	RegEx            string `json:"RegEx"`
	UseLatestVersion bool   `json:"UseLatestVersion"`
}

func (vc *VersionCheck) BuildRequest() (url string, response string) {
	githubInfos, isGithub := strings.CutPrefix(vc.Url, fmt.Sprint(GITHUB_PREFIX, ":"))
	if !isGithub {
		url = vc.Url
	} else {
		owner, repo, ok := strings.Cut(githubInfos, "/")
		if ok {
			url = GITHUB_GRAPHQL_URL
			response = fmt.Sprint(`{"query": "query{repository(owner:\"`, owner, `\", name:\"`, repo, `\") {latestRelease{tagName}}}"}`)
		}
	}

	return
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
