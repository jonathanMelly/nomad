package main

import (
	"github.com/gologme/log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	log.EnableLevelsByNumber(10)
	now := time.Now()
	cwd, _ := os.Getwd()
	target := filepath.Join(cwd, "date.txt")

	err := os.WriteFile(target, []byte(now.Format("2006-06-02 15:04:05")), os.ModePerm)
	if err != nil {
		log.Errorln("Cannot print", now, "to", target, err)
		return
	}
	log.Println(now, "->", target)
	return

}
