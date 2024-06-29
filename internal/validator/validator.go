package validator

import (
	"regexp"
	"strings"
)

var EamilRx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if ok {
		return
	}

	v.AddError(key, message)
}

// -------------------

func In(value string, list ...string) bool {
	for _, s := range list {
		if value == s {
			return true
		}
	}

	return false
}

func Matches(rx regexp.Regexp, test string) bool {
	return rx.MatchString(test)
}

func Unique(values []string) bool {
	seen := make(map[string]bool)

	for _, value := range values {
		normalizedValue := strings.ToLower(value)

		if _, ok := seen[normalizedValue]; ok {
			return false
		}

		seen[normalizedValue] = true
	}

	return true
}
