package version

import (
	"errors"
	"fmt"
	"github.com/gologme/log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Text       string
	Major      *uint
	Minor      *uint
	Patch      *uint
	Patch2     *uint
	Prerelease string
	Build      string
}

// VERSION_REGEX inspired from https://semver.org/
const VERSION_REGEX = `(?P<full>` +
	`(?P<major>0|[1-9]\d*)` +
	`(?:\.(?P<minor>0|[1-9]\d*))?` +
	`(?:\.(?P<patch>0|[1-9]\d*))?` +
	`(?:\.(?P<patch2>0|[1-9]\d*))?` +
	`(?:[.-](?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+(?P<build>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?` +
	`)`

const VERSION_PLACEHOLDER = "{{VERSION}}"

func (version Version) IsNewerThan(other Version) bool {

	log.Traceln("Comparing ", version, " to ", other)

	return version.Major != nil && int(*version.Major) > other.safeGetIntProperty("Major") ||
		version.Minor != nil && int(*version.Minor) > other.safeGetIntProperty("Minor") ||
		version.Patch != nil && int(*version.Patch) > other.safeGetIntProperty("Patch") ||
		version.Patch2 != nil && int(*version.Patch2) > other.safeGetIntProperty("Patch2") ||
		strings.Compare(version.Prerelease, other.Prerelease) == 1 ||
		strings.Compare(version.Build, other.Build) == 1
}

func (version Version) safeGetIntProperty(propertyName string) int {
	rversion := reflect.ValueOf(version)
	rattribute := rversion.FieldByName(propertyName)

	if !rattribute.IsZero() {
		attributeVal := reflect.Indirect(rattribute)
		if attributeVal.Type().ConvertibleTo(reflect.TypeOf(*version.Major)) {
			value := attributeVal.Uint()
			return int(value)
		}
	}

	return -1
}

func buildVersionRegex(regex string) *regexp.Regexp {
	regex = strings.Replace(regex, VERSION_PLACEHOLDER, VERSION_REGEX, -1)
	re, err := regexp.Compile(regex)
	if err != nil {
		log.Error("Cannot build regexp %w", err)
		return nil
	}

	return re
}

func FromStringCustom(source string, regex string) (*Version, error) {
	version := new(Version)

	re := buildVersionRegex(regex)
	matches := re.FindStringSubmatch(source)

	version.Text = getTextPart(matches, re, "full")

	versionPart, err := getUintPart(matches, re, "major")
	if err != nil {
		return nil, err
	}
	if versionPart == nil {
		return nil, errors.New(fmt.Sprint("No version info found for regex ", re.String(), " (", regex, ")", " in "+source))
	}
	version.Major = versionPart

	versionPart, err = getUintPart(matches, re, "minor")
	if err != nil {
		return nil, err
	}
	version.Minor = versionPart

	versionPart, err = getUintPart(matches, re, "patch")
	if err != nil {
		return nil, err
	}
	version.Patch = versionPart

	versionPart, err = getUintPart(matches, re, "patch2")
	if err != nil {
		return nil, err
	}
	version.Patch2 = versionPart

	version.Prerelease = getTextPart(matches, re, "prerelease")
	version.Build = getTextPart(matches, re, "build")

	return version, nil
}

func FromString(source string) (*Version, error) {
	return FromStringCustom(source, VERSION_PLACEHOLDER)
}

func getTextPart(matches []string, re *regexp.Regexp, partName string) string {
	index := re.SubexpIndex(partName)
	if index > 0 && len(matches) > index {
		return matches[index]
	}
	return ""
}

func getUintPart(matches []string, re *regexp.Regexp, partName string) (*uint, error) {
	subexpIndex := re.SubexpIndex(partName)
	if subexpIndex < 0 {
		return nil, errors.New(partName + " part not found for regex '" + re.String() + "'")
	}

	if len(matches) > subexpIndex {
		part := matches[subexpIndex]
		if part != "" {
			asInt, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			result := uint(asInt)
			return &result, nil
		} else {
			return nil, nil
		}
	} else {
		//If part not found, return empty
		return nil, nil
	}
}

func (version Version) String() string {
	return version.Text
}

func (version Version) FillVersionsPlaceholders(input string) string {
	if input == "" {
		return input
	}

	major := ""
	if version.Major != nil {
		major = fmt.Sprint(*version.Major)
	}

	minor := ""
	if version.Minor != nil {
		minor = fmt.Sprint(*version.Minor)
	}

	patch := ""
	if version.Patch != nil {
		patch = fmt.Sprint(*version.Patch)
	}

	patch2 := ""
	if version.Patch2 != nil {
		patch2 = fmt.Sprint(*version.Patch2)
	}

	var versionReplaces = map[string]string{
		"VERSION":        fmt.Sprint(version),
		"VERSION_NO_DOT": strings.ReplaceAll(version.String(), ".", ""),
		"V_MAJOR":        major,
		"V_MINOR":        minor,
		"V_PATCH":        patch,
		"V_PATCH2":       patch2,
		"V_PRERELEASE":   version.Prerelease,
		"V_BUILD":        version.Build,
	}

	for source, replacement := range versionReplaces {
		input = strings.Replace(input, "{{"+source+"}}", replacement, -1)
	}

	return input
}
