package installer

import (
	"io/fs"
	"os"
	"regexp"
	"testing"
)

const sub1 = "sub1"

func Test_guessDeepestRootFolder(t *testing.T) {
	type args struct {
		fsys fs.FS
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "rootAtRootLevel", args: args{fsys: os.DirFS("../../../test/data/notdeepest")}, want: ".", wantErr: false},
		{name: "rootAtSubLevel", args: args{os.DirFS("../../../test/data/deepest")}, want: sub1, wantErr: false},
		{name: "rootAtSubSubLevel", args: args{os.DirFS("../../../test/data/deepestmulti")}, want: "deepestmultisub/deepend/sub", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := guessDeepestRootFolder(tt.args.fsys)
			if (err != nil) != tt.wantErr {
				t.Errorf("guessDeepestRootFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("guessDeepestRootFolder() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_copy(t *testing.T) {
	const target = "../../../test/data/archiver-test-target"
	type args struct {
		sourceFileSystem fs.FS
		root             string
		targetDirectory  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "rootInRoot", args: args{os.DirFS("../../../test/data/notdeepest"), ".", target}, wantErr: false},
		{name: "rootInSub", args: args{os.DirFS("../../../test/data/deepest"), sub1, target}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allRe, _ := regexp.Compile("(.*)")
			if err := copyFromFS(tt.args.sourceFileSystem, tt.args.root, tt.args.targetDirectory, allRe); (err != nil) != tt.wantErr {
				t.Errorf("copyFromFS() error = %v, wantErr %v", err, tt.wantErr)
			}
			err := os.RemoveAll(target)
			if err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}
