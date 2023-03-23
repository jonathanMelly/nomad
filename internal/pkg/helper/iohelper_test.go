package helper

import "testing"

func TestFileOrDirExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"not exist", args{"404"}, false},
		{"not exist", args{"apps/archives/freefilesync-12.1.php?tx=zzz"}, false},
		{"exist", args{"iohelper_test.go"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileOrDirExists(tt.args.filename); got != tt.want {
				t.Errorf("FileOrDirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
