package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	usernameRegex                = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._]{0,22}[A-Za-z0-9])?$`)
	consecutiveSpecialCharsRegex = regexp.MustCompile(`[-_.]{2,}`)
)

// ValidateUsername ensures the username meets the defined criteria
func ValidateUsername(username string) error {
	switch {
	case len(username) < 3 || len(username) > 24:
		return fmt.Errorf("username must be between 3 and 24 characters")
	case strings.ContainsRune(username, ' '):
		return fmt.Errorf("username must not contain spaces")
	case !usernameRegex.MatchString(username):
		return fmt.Errorf("username must match the required pattern")
	case consecutiveSpecialCharsRegex.MatchString(username):
		return fmt.Errorf("username must not contain consecutive special characters")
	}
	return nil
}
