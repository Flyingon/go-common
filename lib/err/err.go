package err

import "fmt"

// GenError ...
func GenError(errPrefix string, msg string) error {
	return fmt.Errorf("%s %s", errPrefix, msg)
}
