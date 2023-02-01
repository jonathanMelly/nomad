package installer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertOne(t *testing.T, input string, expected string) {
	version, error := findVersion(input)
	assert.Nil(t, error)
	assert.Equal(t, expected, version)
}

func TestVersionPatterns(t *testing.T) {

	t.Run("MajorMinorPatchWithNoise", func(t *testing.T) {
		assertOne(t, "hello1.2.3bob", "1.2.3")
	})

	t.Run("Major.Minor.Patch.Patch2", func(t *testing.T) {
		assertOne(t, "1.2.10.6", "1.2.10.6")
	})

	t.Run("Major.Minor.Patch-Prerelease", func(t *testing.T) {
		assertOne(t, "1.2.10-beta", "1.2.10-beta")
	})

	t.Run("Major.Minor.Patch-Prerelease+Build", func(t *testing.T) {
		assertOne(t, "1.2.10-beta+665", "1.2.10-beta+665")
	})

	t.Run("Major.Minor.Patch.Patch2-Prerelease+Build", func(t *testing.T) {
		assertOne(t, "1.2.10.3-beta+665", "1.2.10.3-beta+665")
	})

}
