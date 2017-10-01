package env

import (
	"os"
	"strings"
)

var (
	ENCODING_ENV_UPPERCASE = false
	ENCODING_ENV_PREFIX    = ""
)

func init() {

	upper_str := os.Getenv("ENCODING_ENV_UPPERCASE")
	ENCODING_ENV_UPPERCASE = true &&
		len(upper_str) > 0 &&
		upper_str != "0" &&
		strings.ToUpper(upper_str) != "FALSE" &&
		strings.ToUpper(upper_str) != "F" &&
		strings.ToUpper(upper_str) != "N"

	ENCODING_ENV_PREFIX = os.Getenv("ENCODING_ENV_PREFIX")
}
