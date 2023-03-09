package installer

import (
	"github.com/gookit/goutil/testutil/assert"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T) {
	wd, _ := os.Getwd()
	absTestDir, _ := filepath.Abs(filepath.Join(wd, "../installer_test"))
	tempDir, _ := os.MkdirTemp(absTestDir, "testrun*")
	absTempDir, _ := filepath.Abs(tempDir)

	t.Cleanup(func() {
		os.Chdir("../")
		if err := os.RemoveAll(absTempDir); err != nil {
			assert.Fail(t, "Cannot clean test dir"+absTempDir, err)
		}
	})
	os.Chdir(tempDir)
	here, _ := os.Getwd()
	log.Println("Working in " + here)
}
