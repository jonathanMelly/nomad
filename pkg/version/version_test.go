package version

import (
	"fmt"
	"github.com/gookit/goutil/testutil/assert"
	"log"
	"os"
	"testing"
)

func assertOne(t *testing.T, input string, prependRegex string, expected string) {
	version, error := FromStringCustom(input, prependRegex+VERSION_PLACEHOLDER)
	assert.Nil(t, error)
	assert.Equal(t, expected, version.String())
}

func assertVersionComparison(t *testing.T, v2 string, v1 string) {
	newer, err := FromString(v2)
	if err != nil {
		assert.Fail(t, "Cannot extract version from "+v2, err)
	}
	older, err := FromString(v1)
	if err != nil {
		assert.Fail(t, "Cannot extract version from "+v1, err)
	}

	assert.True(t, newer.IsNewerThan(*older), fmt.Sprint("Version ", newer, " should be reported as newer than ", older))
	assert.False(t, older.IsNewerThan(*newer), fmt.Sprint("Version ", older, " shouldn’t be reported as newer than ", newer))
}

func TestGitReleaseParsing(t *testing.T) {

	// Read entire iohelper content, giving us little control but
	// making it very simple. No need to close the iohelper.
	content, err := os.ReadFile("../../test/data/git-release.json")
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := string(content)

	version, err := FromStringCustom(text, `"tag_name": "v`+VERSION_PLACEHOLDER+`"`)
	if err != nil {
		assert.Fail(t, "Cannot extract version", err)
	}

	assert.Equal(t, "2.39.2.windows.1 2 39 2", version.FillVersionsPlaceholders("{{VERSION}} {{V_MAJOR}} {{V_MINOR}} {{V_PATCH}}"))

}

func TestPlaceholderReplacement(t *testing.T) {

	version, _ := FromString("1.2.3")
	assert.Equal(t, "salut1.2.3 1 2 3", version.FillVersionsPlaceholders("salut{{VERSION}} {{V_MAJOR}} {{V_MINOR}} {{V_PATCH}}"))

}

func TestVersionComparisons(t *testing.T) {

	assets := map[string]string{
		"2.0":        "1.0",
		"1.2.3":      "1.2.2",
		"0.2":        "0.1",
		"1.2.3-beta": "1.2.3-alpha",
		"2":          "1",
		"1.2.9":      "1.2"}
	for new, old := range assets {
		t.Run(new+">"+old, func(t *testing.T) {
			assertVersionComparison(t, new, old)
		})
	}

}

func TestVersionDetailsPatterns(t *testing.T) {

	t.Run("MajorMinorPatchWithNoise", func(t *testing.T) {
		assertOne(t, "hello1.2.3bob", "", "1.2.3")
	})

	t.Run("Major.Minor.Patch.Patch2", func(t *testing.T) {
		assertOne(t, "1.2.10.6", "", "1.2.10.6")
	})

	t.Run("Major.Minor.Patch-Prerelease", func(t *testing.T) {
		assertOne(t, "1.2.10-beta", "", "1.2.10-beta")
	})

	t.Run("Major.Minor.Patch-Prerelease+Build", func(t *testing.T) {
		assertOne(t, "1.2.10-beta+665", "", "1.2.10-beta+665")
	})

	t.Run("Major.Minor.Patch-Prerelease", func(t *testing.T) {
		assertOne(t, "1.2.3-alpha", "", "1.2.3-alpha")
	})

	t.Run("Major.Minor.Patch.Patch2-Prerelease+Build", func(t *testing.T) {
		assertOne(t, "1.2.10.3-beta+665", "", "1.2.10.3-beta+665")
	})

	t.Run("github json", func(t *testing.T) {
		assertOne(t, `"node_id": "RE_kwDOAQ-n9M4FMTDq",
			"tag_name": "v1.61.1",
			"target_commitish": "master"`, `"tag_name": "v`, "1.61.1")
	})

	t.Run("github json", func(t *testing.T) {
		assertOne(t, `"tag_name": "v2.39.2.windows.1"`, `"tag_name": "v`, "2.39.2.windows.1")
	})

}

func TestVersionPatterns(t *testing.T) {
	assets := []string{"29.0.0", "29.0.2"}
	for _, version := range assets {
		t.Run(version, func(t *testing.T) {
			assertOne(t, version, "", version)
		})
	}
}

func TestSafeGetIntPart(t *testing.T) {
	version, _ := FromString("1.2.3.4-alpha+45")
	assert.Equal(t, 1, version.safeGetIntProperty("Major"))
	assert.Equal(t, 2, version.safeGetIntProperty("Minor"))
	assert.Equal(t, 3, version.safeGetIntProperty("Patch"))
}
