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

func TestVersionCheck_Parse(t *testing.T) {
	type fields struct {
		Url              string
		RegEx            string
		UseLatestVersion bool
	}
	tests := []struct {
		name            string
		fields          fields
		wantUrl         string
		wantRequestBody string
	}{
		{"github graphql",
			fields{"github:owner/repo", "", true},
			"https://api.github.com/graphql",
			`{"query": "query{repository(owner:\"owner\", name:\"repo\") {latestRelease{tagName}}}"}`,
		},
		{"standard url",
			fields{"standard", "", true},
			"standard",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc := &VersionCheck{
				Url:              tt.fields.Url,
				RegEx:            tt.fields.RegEx,
				UseLatestVersion: tt.fields.UseLatestVersion,
			}
			gotUrl, gotRequestBody := vc.Parse()
			if gotUrl != tt.wantUrl {
				t.Errorf("Parse() gotUrl = %v, want %v", gotUrl, tt.wantUrl)
			}
			if gotRequestBody != tt.wantRequestBody {
				t.Errorf("Parse() gotRequestBody = %v, want %v", gotRequestBody, tt.wantRequestBody)
			}
		})
	}
}
