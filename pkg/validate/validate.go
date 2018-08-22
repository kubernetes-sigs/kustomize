package validate

import (
	"fmt"
	"regexp"
)

// TODO: these are rudimentary placeholder validation functions and need
// additional work to truly match expected syntax rules.

// IsValidLabel checks whether a label key/value pair has correct syntax and
// character set
func IsValidLabel(keyval string) (bool, error) {
	ok, err := regexp.MatchString(`\A([a-zA-Z0-9_.-]+):([a-zA-Z0-9_.-]+)\z`, keyval)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("invalid label format: %s", keyval)
	}
	return true, nil
}

// IsValidAnnotation checks whether an annotation key/value pair has correct
// syntax and character set
func IsValidAnnotation(keyval string) (bool, error) {
	ok, err := regexp.MatchString(`\A([a-zA-Z0-9_.-]+):([a-zA-Z0-9_.-]+)\z`, keyval)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("invalid annotation format: %s", keyval)
	}
	return true, nil
}
