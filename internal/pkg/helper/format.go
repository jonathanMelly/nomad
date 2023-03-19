package helper

import "fmt"

func BuildPrefix(app string) string {
	return fmt.Sprint("|", app, "| ")
}
