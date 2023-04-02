package data

import (
	"github.com/gookit/goutil/testutil/assert"
	"reflect"
	"testing"
)

func TestAppDefinition_ComputeDownloadExtension(t *testing.T) {
	type fields AppDefinition
	tests := []struct {
		name   string
		fields fields
		result string
	}{
		{name: "custom", fields: fields{DownloadUrl: "test.zip?bob"}, result: ".zip"},
		{name: "standard", fields: fields{DownloadUrl: "http://www.test.com/test.zip"}, result: ".zip"},
		{name: "githubrepo", fields: fields{DownloadUrl: "test-{{VERSION}}.exe"}, result: ".exe"},
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
			gotUrl, gotRequestBody := vc.BuildRequest()
			if gotUrl != tt.wantUrl {
				t.Errorf("BuildRequest() gotUrl = %v, want %v", gotUrl, tt.wantUrl)
			}
			if gotRequestBody != tt.wantRequestBody {
				t.Errorf("BuildRequest() gotRequestBody = %v, want %v", gotRequestBody, tt.wantRequestBody)
			}
		})
	}
}

func TestAppDefinition_fillInfosFromRepository(t *testing.T) {
	type want struct {
		DownloadUrl  string
		VersionCheck VersionCheck
	}
	type fields AppDefinition

	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{"github", fields{RepositoryUrl: "github:owner/repo", DownloadUrl: "test.zip"},
			want{DownloadUrl: "https://github.com/owner/repo/releases/download/test.zip",
				VersionCheck: VersionCheck{Url: "github:owner/repo", RegEx: `"tagName":"[^\d]*{{VERSION}}"`}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := &AppDefinition{
				Version:           tt.fields.Version,
				DownloadUrl:       tt.fields.DownloadUrl,
				RepositoryUrl:     tt.fields.RepositoryUrl,
				ApplicationName:   tt.fields.ApplicationName,
				DownloadExtension: tt.fields.DownloadExtension,
				VersionCheck:      tt.fields.VersionCheck,
				Symlink:           tt.fields.Symlink,
				Shortcut:          tt.fields.Shortcut,
				ShortcutIcon:      tt.fields.ShortcutIcon,
				ExtractRegExList:  tt.fields.ExtractRegExList,
				CreateFolders:     tt.fields.CreateFolders,
				CreateFiles:       tt.fields.CreateFiles,
				MoveObjects:       tt.fields.MoveObjects,
				RestoreFiles:      tt.fields.RestoreFiles,
				validated:         tt.fields.validated,
				extractRegex:      tt.fields.extractRegex,
			}
			definition.fillInfosFromRepository([]string{})
			assert.Equal(t, tt.want.DownloadUrl, definition.DownloadUrl)
			assert.True(t, reflect.DeepEqual(tt.want.VersionCheck, definition.VersionCheck))
		})
	}
}
