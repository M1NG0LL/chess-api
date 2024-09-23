package account

import (
	"regexp"
)

// To verify email
func isValidEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}

func ValidatePassword(password string) (bool, string) {

	// At least one lowercase letter
	lowercaseRegex := regexp.MustCompile(`[a-z]`)
	// At least one uppercase letter
	uppercaseRegex := regexp.MustCompile(`[A-Z]`)
	// At least one digit
	digitRegex := regexp.MustCompile(`\d`)
	// At least one special character
	specialCharRegex := regexp.MustCompile(`[\W_]`)
	// Minimum 8 characters
	lengthRegex := regexp.MustCompile(`.{8,}`)

	if !lowercaseRegex.MatchString(password) {
		return true, "password must contain at least one lowercase letter"
	}
	if !uppercaseRegex.MatchString(password) {
		return true, "password must contain at least one uppercase letter"
	}
	if !digitRegex.MatchString(password) {
		return true, "password must contain at least one digit"
	}
	
	if !specialCharRegex.MatchString(password) {
		return true, "password must contain at least one special character"
	}
	if !lengthRegex.MatchString(password) {
		return true, "password must be at least 8 characters long"
	}
	return false, ""
}