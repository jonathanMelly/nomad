package main

import (
	"github.com/gologme/log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	log.EnableLevelsByNumber(10)
	cwd, _ := os.Getwd()
	target := filepath.Join(cwd, "date.txt")
	swissLoc, _ := time.LoadLocation("Europe/Zurich")

	datetimeStr := time.Now().In(swissLoc).Format("2006-01-02 15:04:05")
	err := os.WriteFile(target, []byte(datetimeStr), os.ModePerm)
	if err != nil {
		log.Errorln("Cannot print", datetimeStr, "to", target, err)
		return
	}
	log.Println(datetimeStr, "->", target)
	return

}
