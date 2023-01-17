package installer

import (
	//"regexp"
	"log"
	"os"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestHelloName(t *testing.T) {

	file, _ := os.Stat(`C:\Users\jonmelly\test\portable\apps\archives`)
	log.Println(file.Size())

	Run("../app-definitions/test.json", "", false, true, "", "archives")
	/*name := "Gladys"
	want := regexp.MustCompile(`\b` + name + `\b`)
	msg, err := Hello("Gladys")
	if !want.MatchString(msg) || err != nil {
		t.Fatalf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	}
	*/
}
