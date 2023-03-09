package installer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// combineRegex will take a string array of regular expressions and compile them
// into a single regular expressions
func combineRegex(s []string) (*regexp.Regexp, error) {
	joined := strings.Join(s, "|")

	re, err := regexp.Compile(joined)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func isWindowsPlatform() bool {
	return strings.Contains(runtime.GOOS, "windows")
}

// https://stackoverflow.com/questions/32438204/create-a-windows-shortcut-lnk-in-go
func createShortcut(linkName string, target string, arguments string, workingDirectory string, description string, destination string, icon string) {

	if isWindowsPlatform() {
		var scriptTxt bytes.Buffer
		scriptTxt.WriteString("option explicit\n\n")
		scriptTxt.WriteString("sub CreateShortCut()\n")
		scriptTxt.WriteString("dim objShell, targetDirectory, objLink\n")
		scriptTxt.WriteString("set objShell = CreateObject(\"WScript.Shell\")\n")
		//scriptTxt.WriteString("targetDirectory = objShell.SpecialFolders(\"")
		scriptTxt.WriteString("targetDirectory = \"")
		scriptTxt.WriteString(filepath.FromSlash(destination))
		//scriptTxt.WriteString("\")\n")
		scriptTxt.WriteString("\"\n")
		scriptTxt.WriteString("set objLink = objShell.CreateShortcut(targetDirectory & \"\\")
		scriptTxt.WriteString(linkName)
		scriptTxt.WriteString(".lnk\")\n")
		scriptTxt.WriteString("objLink.Arguments = \"")
		scriptTxt.WriteString(arguments)
		scriptTxt.WriteString("\"\n")
		scriptTxt.WriteString("objLink.Description = \"")
		scriptTxt.WriteString(description)
		scriptTxt.WriteString("\"\n")
		scriptTxt.WriteString("objLink.TargetPath = \"")
		scriptTxt.WriteString(filepath.FromSlash(target))
		scriptTxt.WriteString("\"\n")

		if icon != "" {
			scriptTxt.WriteString("objLink.IconLocation  = \"")
			scriptTxt.WriteString(filepath.FromSlash(icon))
			scriptTxt.WriteString("\"\n")
		}

		scriptTxt.WriteString("objLink.WindowStyle = 1\n")
		scriptTxt.WriteString("objLink.WorkingDirectory = \"")
		scriptTxt.WriteString(filepath.FromSlash(workingDirectory))
		scriptTxt.WriteString("\"\n")
		scriptTxt.WriteString("objLink.Save\nend sub\n\n")
		scriptTxt.WriteString("call CreateShortCut()")
		//fmt.Println(scriptTxt.String())

		filename := fmt.Sprintf("lnkTo%s.vbs", linkName)
		os.WriteFile(filename, scriptTxt.Bytes(), 0777)
		cmd := exec.Command("wscript", filename)
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
		} else {
			log.Println("Shortcut " + linkName + " generated/updated")
			cmd.Wait()
			os.Remove(filename)
		}

	} else {
		os.Symlink(target, path.Join(destination, linkName))
	}

}
