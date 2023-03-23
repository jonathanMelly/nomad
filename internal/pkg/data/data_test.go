package data

import (
	"github.com/gookit/goutil/testutil/assert"
	"regexp"
	"testing"
)

func TestAppDefinition_ComputeDownloadExtension(t *testing.T) {
	type fields struct {
		ApplicationName   string
		DownloadExtension string
		Version           string
		VersionCheck      VersionCheck
		Symlink           string
		Shortcut          string
		ShortcutIcon      string
		DownloadUrl       string
		ExtractRegExList  []string
		CreateFolders     []string
		CreateFiles       map[string]string
		MoveObjects       map[string]string
		RestoreFiles      []string
		Validated         bool
		ExtractRegex      *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields fields
		result string
	}{
		{name: "custom", fields: fields{DownloadUrl: "test.zip?bob"}, result: ".zip"},
		{name: "standard", fields: fields{DownloadUrl: "http://www.test.com/test.zip"}, result: ".zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := &AppDefinition{
				ApplicationName:   tt.fields.ApplicationName,
				DownloadExtension: tt.fields.DownloadExtension,
				Version:           tt.fields.Version,
				VersionCheck:      tt.fields.VersionCheck,
				Symlink:           tt.fields.Symlink,
				Shortcut:          tt.fields.Shortcut,
				ShortcutIcon:      tt.fields.ShortcutIcon,
				DownloadUrl:       tt.fields.DownloadUrl,
				ExtractRegExList:  tt.fields.ExtractRegExList,
				CreateFolders:     tt.fields.CreateFolders,
				CreateFiles:       tt.fields.CreateFiles,
				MoveObjects:       tt.fields.MoveObjects,
				RestoreFiles:      tt.fields.RestoreFiles,
				Validated:         tt.fields.Validated,
				ExtractRegex:      tt.fields.ExtractRegex,
			}
			definition.ComputeDownloadExtension()
			assert.Equal(t, tt.result, definition.DownloadExtension)
		})
	}
}
