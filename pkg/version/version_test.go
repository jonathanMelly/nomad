package version

import (
	"github.com/gologme/log"
	"github.com/gookit/goutil/testutil/assert"
	"os"
	"reflect"
	"testing"
)

func assertOne(t *testing.T, input string, prependRegex string, expected string) {
	version, err := FromStringCustom(input, prependRegex+VERSION_PLACEHOLDER)
	assert.Nil(t, err)
	assert.Equal(t, expected, version.String())
}

func TestGitReleaseParsing(t *testing.T) {

	// Read entire helper content, giving us little control but
	// making it very simple. No need to close the helper.
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

func TestVersionDetailsPatterns(t *testing.T) {

	t.Run("MajorMinorPatchWithNoise", func(t *testing.T) {
		assertOne(t, "hello1.2.3bob", "", "1.2.3bob")
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

	t.Run("openssh beta version", func(t *testing.T) {
		assertOne(t, `"tag_name": "v9.2.0.0p1-Beta"`, `"tag_name": "(?:[vV])`, "9.2.0.0p1-Beta")
	})

	t.Run("github graphql", func(t *testing.T) {
		assertOne(t, `{"data":{"repository":{"latestRelease":{"tagName":"version bob 1.4.0"}}}}`, `"tagName":"[^\d]*`, "1.4.0")
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

func TestEquality(t *testing.T) {
	assets := []string{"0.78", "0.79"}
	for _, version := range assets {
		t.Run(version, func(t *testing.T) {
			v1, _ := FromString(version)
			v2, _ := FromString(version)
			assert.True(t, reflect.DeepEqual(v1, v2), v1, " not == to ", v2)
		})
	}
}

func TestVersion_IsNewerThan(t *testing.T) {
	log.EnableLevelsByNumber(10)

	tests := []struct {
		name  string
		newer string
		older string
		want  bool
	}{
		{"bug1", "12.1", "11.28", true},
		{"", "2.0", "1.0", true},
		{"", "1.2.3", "1.2.2", true},
		{"", "0.2", "0.1", true},
		{"", "1.2.3-beta", "1.2.3-alpha", true},
		{"", "2", "1", true},
		{"", "1.2.9", "1.2", true},
		{"", "12.1", "11.28", true},
		{"", "12.1.99", "12.1.98", true},
		{"", "12.1.98", "12.1.102", false},
		{"", "12.1.98u2", "12.1.98", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			v1, _ := FromString(tt.newer)
			v2, _ := FromString(tt.older)
			if got := v1.IsNewerThan(v2); got != tt.want {
				t.Errorf("%vIsNewerThan(%v) = %v, want %v", v1, v2, got, tt.want)
			}
			if got := !v2.IsNewerThan(v1); got != tt.want {
				t.Errorf("NOT IsNewerThan() = %v, want %v", got, tt.want)
			}
		})
	}
}
