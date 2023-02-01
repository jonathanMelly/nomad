package installer

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major      *uint
	Minor      *uint
	Patch      *uint
	Patch2     *uint
	Prerelease string
	Build      string
}

// VersionRegex inspired from https://semver.org/
const VersionRegex = `(?P<major>0|[1-9]\d*)` +
	`(?:\.(?P<minor>0|[1-9]\d*))?` +
	`(?:\.(?P<patch>0|[1-9]\d*))?` +
	`(?:\.(?P<patch2>0|[1-9]\d*))?` +
	`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+(?P<build>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`

func FromString(source string) (*Version, error) {
	version := new(Version)

	re, err := regexp.Compile(VersionRegex)
	if err != nil {
		return nil, err
	}

	matches := re.FindStringSubmatch(source)

	versionPart, err := GetUintPart(matches, re, "major")
	if err != nil {
		return nil, err
	}
	if versionPart == nil {
		return nil, errors.New("No version info found in " + source)
	}
	version.Major = versionPart

	versionPart, err = GetUintPart(matches, re, "minor")
	if err != nil {
		return nil, err
	}
	version.Minor = versionPart

	versionPart, err = GetUintPart(matches, re, "patch")
	if err != nil {
		return nil, err
	}
	version.Patch = versionPart

	versionPart, err = GetUintPart(matches, re, "patch2")
	if err != nil {
		return nil, err
	}
	version.Patch2 = versionPart

	version.Prerelease = GetTextPart(matches, re, "prerelease")
	version.Build = GetTextPart(matches, re, "build")

	return version, nil

}

func GetTextPart(matches []string, re *regexp.Regexp, partName string) string {
	index := re.SubexpIndex(partName)
	if len(matches) > index {
		return matches[index]
	}
	return ""
}

func GetUintPart(matches []string, re *regexp.Regexp, partName string) (*uint, error) {
	subexpIndex := re.SubexpIndex(partName)
	if subexpIndex < 0 {
		return nil, errors.New("Unavailable group with name " + partName)
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

	return nil, errors.New("Cannot extract " + partName)
}

func (version Version) ToString(original bool) string {
	result := strings.Builder{}
	validParts := make([]string, 0, 5)

	for _, part := range []*uint{version.Major, version.Minor, version.Patch, version.Patch2} {
		currentLen := len(validParts)
		if part != nil || !original {
			validParts = validParts[0 : currentLen+1]
			if part == nil {
				*part = 0
			}
			validParts[currentLen] = fmt.Sprint(*part)
			currentLen++
		}
	}
	result.WriteString(strings.Join(validParts, "."))
	if version.Prerelease != "" {
		result.WriteString("-")
		result.WriteString(version.Prerelease)
	}

	if version.Build != "" {
		result.WriteString("+")
		result.WriteString(version.Build)
	}

	return result.String()
}

func findVersion(text string) (string, error) {
	version, err := FromString(text)
	if err != nil {
		return "", err
	}
	return version.ToString(true), nil
}

func replaceWithVersion(input string) (string, error) {
	return "", nil
}
