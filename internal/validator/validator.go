package validator

import (
	"regexp"
)

// Declare a regular expression for sanity checking the format of email addresses (we'll use later)
// this regular expression pattern is taken from https://html.spec.whatwg.org/#valid-e-mail-address.
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\. [a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Define a new Validator type which contains a map of validation errors.
type Validator struct { 
	Errors map[string]string
}

// New is a helper which creates a new Validator instance with an empty errors map.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if the errors map doesn't contain any entries.
func (validator *Validator) Valid() bool { 
	return len(validator.Errors) == 0
}

// AddError adds an error message to the map (as long as no entry already exists for the given key).
func (validator *Validator) AddError(key, message string) {
	if _, exists := validator.Errors[key]; !exists { 
		validator.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is not 'ok'.
func (validator *Validator) Check(ok bool, key, message string) { 
	if !ok {
		validator.AddError(key, message)
	}
}

// In returns true if a specific value is in a list of strings.
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] { 
			return true
		}
	}
	return false
}

// Matches returns true if a string value matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool { 
	return rx.MatchString(value)
}

// Unique returns true if all string values in a slice are unique.
func Unique(values []string) bool { uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}

// In the code above we’ve defined a custom Validator type which contains a map of errors.
// The Validator type provides a Check() method for conditionally adding errors to the map,
// and a Valid() method which returns whether the errors map is empty or not.
// We’ve also added In(), Matches() and Unique() functions to help us perform some specific validation checks.